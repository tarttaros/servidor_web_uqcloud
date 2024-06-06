package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/user"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

func GestionContenedores(c *gin.Context) {
	// Renderiza la plantilla HTML

	// Acceder a la sesión
	session := sessions.Default(c)
	email := session.Get("email")

	if email == nil {
		// Si el usuario no está autenticado, redirige a la página de inicio de sesión
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// Recuperar o inicializar un arreglo de máquinas virtuales en la sesión del usuario
	machines, _ := MaquinasActualesC(email.(string))

	c.HTML(http.StatusOK, "gestionContenedores.html", gin.H{
		"email":    email,
		"machines": machines,
	})
}

func CrearContenedor(c *gin.Context) {

	serverURL := "http://servidor_procesamiento:8081/json/crearContenedor"

	// Acceder a la sesión
	session := sessions.Default(c)
	email := session.Get("email")

	MaquinaVM := c.PostForm("selectedMachineContenedor")

	// Dividir la cadena en IP y hostname
	partes := strings.Split(MaquinaVM, " - ")
	if len(partes) != 2 {
		// Manejar un error si el formato no es el esperado
		c.JSON(http.StatusBadRequest, gin.H{"error": "Formato de máquina virtual incorrecto"})
		return
	}

	ip := partes[0]
	hostname := partes[1]

	fmt.Println(MaquinaVM)

	nombreImagen := c.PostForm("buscarImagen")

	comando := "docker run "

	deatch := c.PostForm("hiddenInput1")

	if deatch != "" {
		comando += deatch + " "
	}

	remove := c.PostForm("hiddenInput2")

	if remove != "" {
		comando += remove + " "
	}

	name := c.PostForm("name")

	if name != "" {
		comando += " --name " + "'" + name + "' "
	}

	port := c.PostForm("port")

	if port != "" {
		comando += " -p " + port + " "
	}

	memory := c.PostForm("memory")

	if memory != "" {
		comando += " --memory " + memory + "M "
	}

	volume := c.PostForm("volume")
	fmt.Println("Volume:", volume)

	// Obtener el archivo del formulario
	file, fileHeader, err := c.Request.FormFile("archivo")

	// Verificar si el archivo y el volumen están presentes
	if file != nil && volume != "" {

		if err != nil {
			fmt.Println("Error al obtener el archivo:", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "No se pudo obtener el archivo"})
			return
		}

		defer file.Close()

		usuario := obtenerUsuario()

		// Guardar el archivo temporalmente en el servidor
		archivoTemporal := "/home/" + usuario + "/" + fileHeader.Filename
		err = c.SaveUploadedFile(fileHeader, archivoTemporal)
		if err != nil {
			fmt.Println("Error al guardar el archivo temporalmente:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al guardar el archivo en el servidor"})
			return
		}

		config, err := configurarSSHContraseniaC(hostname)

		if err != nil {
			fmt.Println("Error al configurar SSH:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al configurar SSH"})
			return
		}

		partes = strings.Split(archivoTemporal, "/")
		archivo := partes[len(partes)-1]
		fmt.Println("Archivo:", archivo)

		rutaArchivo, err := enviarArchivo(ip, archivoTemporal, archivo, hostname, config)

		if err != nil {
			fmt.Println("Error al enviar el archivo:", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Error al enviar el archivo"})
			return
		}

		fmt.Println("Ruta del archivo enviado:", rutaArchivo)

		rutaCarpeta := desmpacetarArchivo(config, archivo, hostname, ip)

		comando += " -v " + rutaCarpeta + ":" + volume + " "

		err = os.Remove(archivoTemporal)
		if err != nil {
			// Manejar el error si no se puede eliminar el archivo temporal
			log.Println("Error al eliminar el archivo temporal:", err)
		}

	} else if file == nil && volume != "" {
		// Caso donde no hay archivo pero sí volumen
		fmt.Println("Ingresando sin archivo pero con volumen")

		// Lógica para manejar el caso cuando solo hay volumen
		// Puedes ejecutar tu comando o realizar otras acciones
		comando += " -v " + volume + " "

	}

	fmt.Println(comando)

	payload := map[string]interface{}{
		"imagen":   nombreImagen,
		"comando":  comando,
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
	machines, _ := MaquinasActualesC(email.(string))

	// Renderizar la plantilla HTML con los datos recibidos, incluyendo el mensaje
	c.HTML(http.StatusOK, "gestionContenedores.html", gin.H{
		"mensaje":  mensaje, // Pasar el mensaje al contexto de renderizado
		"email":    email,
		"machines": machines,
	})

}

func CorrerContenedor(c *gin.Context) {

	serverURL := "http://servidor_procesamiento:8081/json/gestionContenedor"

	// Acceder a la sesión
	session := sessions.Default(c)
	email := session.Get("email")

	MaquinaVM := c.PostForm("selectedMachineContenedor")

	fmt.Println(MaquinaVM)

	idContenedor := c.PostForm("IdContenedor")

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
		"solicitud":  "correr",
		"contenedor": idContenedor,
		"ip":         ip,
		"hostname":   hostname,
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
	machines, _ := MaquinasActualesC(email.(string))

	// Renderizar la plantilla HTML con los datos recibidos, incluyendo el mensaje
	c.HTML(http.StatusOK, "gestionContenedores.html", gin.H{
		"mensaje":  mensaje, // Pasar el mensaje al contexto de renderizado
		"email":    email,
		"machines": machines,
	})

}

func PausarContenedor(c *gin.Context) {

	serverURL := "http://servidor_procesamiento:8081/json/gestionContenedor"

	// Acceder a la sesión
	session := sessions.Default(c)
	email := session.Get("email")

	MaquinaVM := c.PostForm("selectedMachineContenedor")

	fmt.Println(MaquinaVM)

	idContenedor := c.PostForm("IdContenedor")

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
		"solicitud":  "pausar",
		"contenedor": idContenedor,
		"ip":         ip,
		"hostname":   hostname,
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
	machines, _ := MaquinasActualesC(email.(string))

	// Renderizar la plantilla HTML con los datos recibidos, incluyendo el mensaje
	c.HTML(http.StatusOK, "gestionContenedores.html", gin.H{
		"mensaje":  mensaje, // Pasar el mensaje al contexto de renderizado
		"email":    email,
		"machines": machines,
	})

}

func ReiniciarContenedor(c *gin.Context) {

	serverURL := "http://servidor_procesamiento:8081/json/gestionContenedor"

	// Acceder a la sesión
	session := sessions.Default(c)
	email := session.Get("email")

	MaquinaVM := c.PostForm("selectedMachineContenedor")

	fmt.Println(MaquinaVM)

	idContenedor := c.PostForm("IdContenedor")

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
		"solicitud":  "reiniciar",
		"contenedor": idContenedor,
		"ip":         ip,
		"hostname":   hostname,
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
	machines, _ := MaquinasActualesC(email.(string))

	// Renderizar la plantilla HTML con los datos recibidos, incluyendo el mensaje
	c.HTML(http.StatusOK, "gestionContenedores.html", gin.H{
		"mensaje":  mensaje, // Pasar el mensaje al contexto de renderizado
		"email":    email,
		"machines": machines,
	})

}

func EliminarContenedor(c *gin.Context) {

	serverURL := "http://servidor_procesamiento:8081/json/gestionContenedor"

	// Acceder a la sesión
	session := sessions.Default(c)
	email := session.Get("email")

	MaquinaVM := c.PostForm("selectedMachineContenedor")

	fmt.Println(MaquinaVM)

	idContenedor := c.PostForm("IdContenedor")

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
		"solicitud":  "borrar",
		"contenedor": idContenedor,
		"ip":         ip,
		"hostname":   hostname,
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
	machines, _ := MaquinasActualesC(email.(string))

	// Renderizar la plantilla HTML con los datos recibidos, incluyendo el mensaje
	c.HTML(http.StatusOK, "gestionContenedores.html", gin.H{
		"mensaje":  mensaje, // Pasar el mensaje al contexto de renderizado
		"email":    email,
		"machines": machines,
	})

}

func EliminarContenedores(c *gin.Context) {
	serverURL := "http://servidor_procesamiento:8081/json/gestionContenedor"

	// Acceder a la sesión
	session := sessions.Default(c)
	email := session.Get("email")

	MaquinaVM := c.PostForm("selectedMachineC")

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
		"solicitud": "eliminar",
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

	var respuesta map[string]string

	err = json.NewDecoder(resp.Body).Decode(&respuesta)
	if err != nil {
		log.Println("Error al decodificar el body de la respuesta")
		return
	}

	mensaje := respuesta["mensaje"]

	// Recuperar o inicializar un arreglo de máquinas virtuales en la sesión del usuario
	machines, _ := MaquinasActualesC(email.(string))

	// Renderizar la plantilla HTML con los datos recibidos, incluyendo el mensaje
	c.HTML(http.StatusOK, "gestionContenedores.html", gin.H{
		"mensaje":  mensaje, // Pasar el mensaje al contexto de renderizado
		"email":    email,
		"machines": machines,
	})
}

func GetContendores(c *gin.Context) {

	// Acceder a la sesión para obtener el email del usuario
	maquinaVirtual := c.PostForm("buscarMV")

	log.Println("Maquina Virtual:", maquinaVirtual)

	// Obtener los datos de las máquinas utilizando el email del usuario
	contenedores, err := obtenerContenedores(maquinaVirtual)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Obtener los datos de las máquinas utilizando el email del usuario
	imagen, err := ObtenerImagenesC(maquinaVirtual)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"contenedores": contenedores,
		"imagen":       imagen,
	})

}

func obtenerContenedores(maquinaVirtual string) ([]Conetendor, error) {
	serverURL := "http://servidor_procesamiento:8081/json/ContenedoresVM" // Cambia esto por la URL de tu servidor en el puerto 8081

	partes := strings.Split(maquinaVirtual, " - ")

	ip := partes[0]
	hostname := partes[1]

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
	var contenedores []Conetendor

	// Decodifica los datos de respuesta en la variable machines.
	if err := json.Unmarshal(responseBody, &contenedores); err != nil {
		// Maneja el error de decodificación aquí
	}

	return contenedores, nil

}

func ObtenerImagenesC(maquinaVirtual string) ([]Imagen, error) {
	// Lee la información de la máquina virtual seleccionada del cuerpo de la solicitud

	partes := strings.Split(maquinaVirtual, " - ")

	serverURL := "http://servidor_procesamiento:8081/json/imagenesVM"

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

func MaquinasActualesC(email string) ([]Maquina_virtual, error) {
	serverURL := "http://servidor_procesamiento:8081/json/consultMachine" // Cambia esto por la URL de tu servidor en el puerto 8081

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

func enviarArchivo(host, archivoLocal, nombreArchivo, hostname string, config *ssh.ClientConfig) (salida string, err error) {

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
	remoteFile, err := sftpClient.Create("/home/" + hostname + "/" + nombreArchivo)
	if err != nil {
		log.Fatalf("Failed to create remote file: %v", err)
	}
	defer remoteFile.Close()

	// Escribir el contenido del archivo local en el archivo remoto
	_, err = remoteFile.Write(localFile)
	if err != nil {
		log.Fatalf("Failed to write to remote file: %v", err)
	}

	return "/home/" + hostname + "/" + nombreArchivo, nil

}

func obtenerUsuario() string {
	// Obtiene la información del usuario actual
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	return usr.Username

}

func configurarSSHContraseniaC(user string) (*ssh.ClientConfig, error) {

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

func desmpacetarArchivo(config *ssh.ClientConfig, archivo, hostname, ip string) string {

	partes := strings.Split(archivo, ".")
	nombreCarpeta := partes[0]
	fmt.Println("nombre Carpeta:", nombreCarpeta)

	sctlCommand := "mkdir /home/" + hostname + "/" + nombreCarpeta + "&&" + " unzip " + archivo + " -d /home/" + hostname + "/" + nombreCarpeta

	_, err3 := enviarComandoSSH(ip, sctlCommand, config)

	if err3 != nil {
		log.Println("Error al configurar SSH:", err3)
		return "Error al configurar la conexiòn SSH"
	}

	sctlCommand = "rm " + archivo

	_, err3 = enviarComandoSSH(ip, sctlCommand, config)

	if err3 != nil {
		log.Println("Error al configurar SSH:", err3)
		return "Error al configurar la conexiòn SSH"
	}

	return "/home/" + hostname + "/" + nombreCarpeta

}

func enviarComandoSSH(host string, comando string, config *ssh.ClientConfig) (salida string, err error) {

	//Establece la conexiòn SSH
	conn, err := ssh.Dial("tcp", host+":22", config)
	if err != nil {
		log.Println("Error al establecer la conexiòn SSH: ", err)
		return "", err
	}
	defer conn.Close()

	//Crea una nueva sesiòn SSH
	session, err := conn.NewSession()
	if err != nil {
		log.Println("Error al crear la sesiòn SSH: ", err)
		return "", err
	}
	defer session.Close()
	//Ejecuta el comando remoto
	output, err := session.CombinedOutput(comando)
	if err != nil {
		log.Println("Error al ejecutar el comando remoto: " + string(output))
		return "", err
	}
	return string(output), nil
}
