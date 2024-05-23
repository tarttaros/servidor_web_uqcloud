package handlers

import (
	"bytes"
	"os/user"

	// "context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	// "time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// Esta función maneja la solicitud GET a la ruta /gestionImagenes.
func GestionImagenes(c *gin.Context) {

	// Acceder a la sesión
	session := sessions.Default(c)
	email := session.Get("email")

	if email == nil {
		// Si el usuario no está autenticado, redirige a la página de inicio de sesión
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// Recuperar o inicializar un arreglo de máquinas virtuales en la sesión del usuario
	machines, _ := MaquinasActualesI(email.(string))

	c.HTML(http.StatusOK, "gestionImagenes.html", gin.H{
		"email":    email,
		"machines": machines,
	})
}

func CrearImagen(c *gin.Context) {
	serverURL := "http://localhost:8081/json/imagenHub"

	// Acceder a la sesión
	session := sessions.Default(c)
	email := session.Get("email")

	// Obtener datos del formulario
	nombreImagen := c.PostForm("nombreImagen")
	versionImagen := c.PostForm("versionImagen")

	MaquinaVM := c.PostForm("selectedMachineDockerHub")

	fmt.Println(MaquinaVM)

	// Dividir la cadena en IP y hostname
	partes := strings.Split(MaquinaVM, " - ")
	if len(partes) != 2 {
		// Manejar un error si el formato no es el esperado
		c.JSON(http.StatusBadRequest, gin.H{"error": "Formato de máquina virtual incorrecto"})
		return
	}

	ip := partes[0]
	hostname := partes[1]

	payload := map[string]interface{}{
		"imagen":   nombreImagen,
		"version":  versionImagen,
		"ip":       ip,
		"hostname": hostname,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return
	}

	// Crea una solicitud HTTP POST con el JSON como cuerpo
	req, err := http.NewRequest("POST", serverURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return
	}

	// Establece el encabezado de tipo de contenido
	req.Header.Set("Content-Type", "application/json")

	// Realiza la solicitud HTTP
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var respuesta map[string]string

	err = json.NewDecoder(resp.Body).Decode(&respuesta)
	if err != nil {
		log.Println("Error al decodificar el body de la respuesta")
		return
	}

	mensaje := respuesta["mensaje"]

	// Recuperar o inicializar un arreglo de máquinas virtuales en la sesión del usuario
	machines, _ := MaquinasActualesI(email.(string))

	// Renderizar la plantilla HTML con los datos recibidos, incluyendo el mensaje
	c.HTML(http.StatusOK, "gestionImagenes.html", gin.H{
		"email":    email,
		"mensaje":  mensaje, // Pasar el mensaje al contexto de renderizado
		"machines": machines,
	})
}

func MaquinasActualesI(email string) ([]Maquina_virtual, error) {
	serverURL := "http://localhost:8081/json/consultMachine" // Cambia esto por la URL de tu servidor en el puerto 8081

	persona := Persona{Email: email}
	jsonData, err := json.Marshal(persona)
	if err != nil {
		return nil, err
	}

	// Crea una solicitud HTTP POST con el JSON como cuerpo
	req, err := http.NewRequest("POST", serverURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	// Establece el encabezado de tipo de contenido
	req.Header.Set("Content-Type", "application/json")

	// Realiza la solicitud HTTP
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Verifica la respuesta del servidor (resp.StatusCode) aquí si es necesario
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("La solicitud al servidor no fue exitosa")
	}

	// Lee la respuesta del cuerpo de la respuesta HTTP
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var machines []Maquina_virtual

	// Decodifica los datos de respuesta en la variable machines.
	if err := json.Unmarshal(responseBody, &machines); err != nil {
		// Maneja el error de decodificación aquí
	}

	return machines, nil
}

func CrearImagenArchivoTar(c *gin.Context) {

	serverURL := "http://localhost:8081/json/imagenTar"

	// Acceder a la sesión
	session := sessions.Default(c)
	email := session.Get("email")

	// Obtener datos del formulario
	MaquinaVM := c.PostForm("selectedMachineArchivoTar")
	nombreImagen := c.PostForm("nombreImagen")

	fmt.Println(MaquinaVM)
	// Obtener el archivo del formulario
	file, fileHeader, err := c.Request.FormFile("archivo")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No se pudo obtener el archivo"})
		return
	}
	defer file.Close()

	usuario := obtenerUsuario()

	// Guardar el archivo temporalmente en el servidor
	archivoTemporal := "/home/" + usuario + "/" + fileHeader.Filename
	err = c.SaveUploadedFile(fileHeader, archivoTemporal)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al guardar el archivo en el servidor"})
		return
	}

	// Dividir la cadena en IP y hostname
	partes := strings.Split(MaquinaVM, " - ")
	if len(partes) != 2 {
		// Manejar un error si el formato no es el esperado
		c.JSON(http.StatusBadRequest, gin.H{"error": "Formato de máquina virtual incorrecto"})
		return
	}

	ip := partes[0]
	hostname := partes[1]

	partes = strings.Split(archivoTemporal, "/")

	archivo := partes[len(partes)-1]

	config, err := configurarSSHContrasenia(hostname)

	if err != nil {
		fmt.Println("Error al configurar SSH:", err)
	}

	enviarArchivoSFTP(ip, archivoTemporal, archivo, hostname, config)

	fmt.Println(nombreImagen)

	payload := map[string]interface{}{
		"archivo":      archivo,
		"nombreImagen": nombreImagen,
		"ip":           ip,
		"hostname":     hostname,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return
	}

	// Crea una solicitud HTTP POST con el JSON como cuerpo
	req, err := http.NewRequest("POST", serverURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return
	}

	// Establece el encabezado de tipo de contenido
	req.Header.Set("Content-Type", "application/json")

	// Realiza la solicitud HTTP
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var respuesta map[string]string

	err = json.NewDecoder(resp.Body).Decode(&respuesta)
	if err != nil {
		log.Println("Error al decodificar el body de la respuesta")
		return
	}

	err = os.Remove(archivoTemporal)
	if err != nil {
		// Manejar el error si no se puede eliminar el archivo temporal
		log.Println("Error al eliminar el archivo temporal:", err)
	}

	mensaje := respuesta["mensaje"]

	// Recuperar o inicializar un arreglo de máquinas virtuales en la sesión del usuario
	machines, _ := MaquinasActualesI(email.(string))

	// Renderizar la plantilla HTML con los datos recibidos, incluyendo el mensaje
	c.HTML(http.StatusOK, "gestionImagenes.html", gin.H{
		"email":    email,
		"mensaje":  mensaje, // Pasar el mensaje al contexto de renderizado
		"machines": machines,
	})

}

func configurarSSHContrasenia(user string) (*ssh.ClientConfig, error) {

	fmt.Println("\nconfigurarSSH")

	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password("uqcloud"),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	return config, nil
}

func enviarArchivoSFTP(host, archivoLocal, nombreImagen, hostname string, config *ssh.ClientConfig) (salida string, err error) {

	fmt.Println("\nEnviarArchivos")

	fmt.Println("\n" + host)

	client, err := ssh.Dial("tcp", host+":22", config)
	if err != nil {
		log.Fatalf("Failed to dial: %v", err)
	}
	defer client.Close()

	// Inicializar el cliente SFTP
	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		log.Fatalf("Failed to create SFTP client: %v", err)
	}
	defer sftpClient.Close()

	// Abrir el archivo local
	localFile, err := ioutil.ReadFile(archivoLocal)
	if err != nil {
		log.Fatalf("Failed to read local file: %v", err)
	}

	// Crear el archivo remoto
	remoteFile, err := sftpClient.Create("/home/" + hostname + "/" + nombreImagen)
	if err != nil {
		log.Fatalf("Failed to create remote file: %v", err)
	}
	defer remoteFile.Close()

	// Escribir el contenido del archivo local en el archivo remoto
	_, err = remoteFile.Write(localFile)
	if err != nil {
		log.Fatalf("Failed to write to remote file: %v", err)
	}

	return "Envio Exitoso", nil

}

func CrearImagenDockerFile(c *gin.Context) {

	serverURL := "http://localhost:8081/json/imagenDockerFile"

	// Acceder a la sesión
	session := sessions.Default(c)
	email := session.Get("email")

	// Obtener datos del formulario
	MaquinaVM := c.PostForm("selectedMachineDockerFile")
	nombreImagen := c.PostForm("nombreImagen")

	fmt.Println(MaquinaVM)
	// Obtener el archivo del formulario
	file, fileHeader, err := c.Request.FormFile("archivo")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No se pudo obtener el archivo"})
		return
	}
	defer file.Close()

	usuario := obtenerUsuario()

	// Guardar el archivo temporalmente en el servidor
	archivoTemporal := "/home/" + usuario + "/" + fileHeader.Filename
	err = c.SaveUploadedFile(fileHeader, archivoTemporal)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al guardar el archivo en el servidor"})
		return
	}

	// Dividir la cadena en IP y hostname
	partes := strings.Split(MaquinaVM, " - ")
	if len(partes) != 2 {
		// Manejar un error si el formato no es el esperado
		c.JSON(http.StatusBadRequest, gin.H{"error": "Formato de máquina virtual incorrecto"})
		return
	}

	ip := partes[0]
	hostname := partes[1]

	partes = strings.Split(archivoTemporal, "/")

	archivo := partes[len(partes)-1]

	config, err := configurarSSHContrasenia(hostname)

	if err != nil {
		fmt.Println("Error al configurar SSH:", err)
	}

	enviarArchivoSFTP(ip, archivoTemporal, archivo, hostname, config)

	fmt.Println(nombreImagen)

	payload := map[string]interface{}{
		"archivo":      archivo,
		"nombreImagen": nombreImagen,
		"ip":           ip,
		"hostname":     hostname,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return
	}

	// Crea una solicitud HTTP POST con el JSON como cuerpo
	req, err := http.NewRequest("POST", serverURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return
	}

	// Establece el encabezado de tipo de contenido
	req.Header.Set("Content-Type", "application/json")

	// Realiza la solicitud HTTP
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var respuesta map[string]string

	err = json.NewDecoder(resp.Body).Decode(&respuesta)
	if err != nil {
		log.Println("Error al decodificar el body de la respuesta")
		return
	}

	err = os.Remove(archivoTemporal)
	if err != nil {
		// Manejar el error si no se puede eliminar el archivo temporal
		log.Println("Error al eliminar el archivo temporal:", err)
	}

	mensaje := respuesta["mensaje"]

	// Recuperar o inicializar un arreglo de máquinas virtuales en la sesión del usuario
	machines, _ := MaquinasActualesI(email.(string))

	// Renderizar la plantilla HTML con los datos recibidos, incluyendo el mensaje
	c.HTML(http.StatusOK, "gestionImagenes.html", gin.H{
		"email":    email,
		"mensaje":  mensaje, // Pasar el mensaje al contexto de renderizado
		"machines": machines,
	})

}

func EliminarImagen(c *gin.Context) {

	fmt.Println("Eliminar")

	serverURL := "http://localhost:8081/json/eliminarImagen"

	// Acceder a la sesión
	session := sessions.Default(c)
	email := session.Get("email")

	// Obtener datos del formulario
	nombreImagen := c.PostForm("imagenRepositorio")

	MaquinaVM := c.PostForm("selectedMachineImagen")

	fmt.Println(MaquinaVM)

	// Dividir la cadena en IP y hostname
	partes := strings.Split(MaquinaVM, " - ")
	if len(partes) != 2 {
		// Manejar un error si el formato no es el esperado
		c.JSON(http.StatusBadRequest, gin.H{"error": "Formato de máquina virtual incorrecto"})
		return
	}

	ip := partes[0]
	hostname := partes[1]
	solicitud := "borar"

	payload := map[string]interface{}{
		"imagen":    nombreImagen,
		"solicitud": solicitud,
		"ip":        ip,
		"hostname":  hostname,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return
	}

	// Crea una solicitud HTTP POST con el JSON como cuerpo
	req, err := http.NewRequest("POST", serverURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return
	}

	// Establece el encabezado de tipo de contenido
	req.Header.Set("Content-Type", "application/json")

	// Realiza la solicitud HTTP
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	// Recuperar o inicializar un arreglo de máquinas virtuales en la sesión del usuario
	machines, _ := MaquinasActualesI(email.(string))

	// Renderizar la plantilla HTML con los datos recibidos, incluyendo el mensaje
	c.HTML(http.StatusOK, "gestionImagenes.html", gin.H{
		"machines": machines,
	})
}

func ObtenerImagenes(maquinaVirtual string) ([]Imagen, error) {
	// Lee la información de la máquina virtual seleccionada del cuerpo de la solicitud

	partes := strings.Split(maquinaVirtual, " - ")

	serverURL := "http://localhost:8081/json/imagenesVM"

	ip := partes[0]
	hostname := partes[1]

	fmt.Println(ip + "-" + hostname)

	payload := map[string]interface{}{
		"ip":       ip,
		"hostname": hostname,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err

	}

	// Crea una solicitud HTTP POST con el JSON como cuerpo
	req, err := http.NewRequest("POST", serverURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	// Establece el encabezado de tipo de contenido
	req.Header.Set("Content-Type", "application/json")

	// Realiza la solicitud HTTP
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Verifica la respuesta del servidor (resp.StatusCode) aquí si es necesario
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("La solicitud al servidor no fue exitosa")
	}

	// Lee la respuesta del cuerpo de la respuesta HTTP
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var imagenes []Imagen

	// Decodifica los datos de respuesta en la variable machines.
	if err := json.Unmarshal(responseBody, &imagenes); err != nil {
		// Maneja el error de decodificación aquí
	}

	return imagenes, nil

}

func GetImages(c *gin.Context) {

	// Acceder a la sesión para obtener el email del usuario
	maquinaVirtual := c.PostForm("buscarMV")

	log.Println("Maquina Virtual:", maquinaVirtual)

	// Obtener los datos de las máquinas utilizando el email del usuario
	imagen, err := ObtenerImagenes(maquinaVirtual)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, imagen)

}

func obtenerUsuario() string {
	// Obtiene la información del usuario actual
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	return usr.Username

}
