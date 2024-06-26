package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func ControlMachine(c *gin.Context) {
	// Acceder a la sesión
	session := sessions.Default(c)
	email := session.Get("email")

	if email == nil {
		// Si el usuario no está autenticado, redirige a la página de inicio de sesión
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// Recuperar o inicializar un arreglo de máquinas virtuales en la sesión del usuario
	machines, _ := consultarMaquinas(email.(string))

	hosts, _ := consultarHostDisponibles()

	if sessionMachines, ok := session.Get("machines").([]Maquina_virtual); ok {
		machines = sessionMachines
	} else {
		// Inicializa un nuevo arreglo de máquinas si no existe en la sesión
		machines = []Maquina_virtual{}
	}

	// Agregar la variable booleana `machinesChange` a la sesión y establecerla en true
	session.Set("machinesChange", true)
	session.Save()

	machinesChange := session.Get("machinesChange")
	clientIP := c.ClientIP()
	showNewButton := false
	for _, host := range hosts {
		// Depuración
		if host.Ip == clientIP {
			showNewButton = true
			break
		}
	}
	c.HTML(http.StatusOK, "controlMachine.html", gin.H{
		"email":          email,
		"machines":       machines,
		"machinesChange": machinesChange,
		"hosts":          hosts,
		"showNewButton":  showNewButton,
		"clientIP":       clientIP,
	})
}

func MainSend(c *gin.Context) {
	// Acceder a la sesión
	session := sessions.Default(c)
	email := session.Get("email")

	if email == nil {
		// Si el usuario no está autenticado, redirige a la página de inicio de sesión
		c.Redirect(http.StatusFound, "/login")
		return
	}

	userEmail := email.(string)

	// Obtener los datos del formulario
	vmname := c.PostForm("vmnameCreate")
	if vmname == "" {
		// Si el nombre de la máquina virtual está vacío, mostrar un mensaje de error en el HTML
		c.HTML(http.StatusOK, "controlMachine.html", gin.H{
			"ErrorMessage": "El nombre de la máquina virtual no puede estar vacío.",
		})
		return
	}
	ditOs := c.PostForm("osCreate")
	memoryStr := c.PostForm("memoryCreate")
	memory, err := strconv.Atoi(memoryStr)
	cpuStr := c.PostForm("cpuCreate")
	cpu, _ := strconv.Atoi(cpuStr)
	os := "Linux"

	host := c.PostForm("host")
	hostStr, _ := strconv.Atoi(host) //Variables creadas en la segunda iteracion desktop cloud , Caso de uso : Asignacion de Recursos

	// Crear una estructura Account y convertirla a JSON
	// En la declaracion de esta variable se adiciono el host seleccionado
	maquina_virtual := Maquina_virtual{Nombre: vmname, Sistema_operativo: os, Distribucion_sistema_operativo: ditOs, Ram: memory, Cpu: cpu, Persona_email: userEmail, Host_id: hostStr}
	clientIP := c.ClientIP()

	payload := map[string]interface{}{
		"specifications": maquina_virtual,
		"clientIP":       clientIP,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if sendJSONMachineToServer(jsonData) {
		// Registro exitoso, muestra un mensaje de éxito en el HTML
		c.HTML(http.StatusOK, "controlMachine.html", gin.H{
			"SuccessMessage": "Solicitud para crear màquina virtual enviada con èxito.",
		})
	} else {
		// Registro erróneo, muestra un mensaje de error en el HTML
		c.HTML(http.StatusOK, "controlMachine.html", gin.H{
			"ErrorMessage": "Error al enviar la solicitud para crear màquina virtual. Intente de nuevo",
		})
	}
}

func sendJSONMachineToServer(jsonData []byte) bool {
	serverURL := "http://servidor_procesamiento:8081/json/createVirtualMachine"

	// Crea una solicitud HTTP POST con el JSON como cuerpo
	req, err := http.NewRequest("POST", serverURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return false
	}

	// Establece el encabezado de tipo de contenido
	req.Header.Set("Content-Type", "application/json")

	// Realiza la solicitud HTTP
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	// Verifica la respuesta del servidor (resp.StatusCode) aquí si es necesario
	if resp.StatusCode != http.StatusOK {
		return false
	} else {
		return true
	}
}

func consultarMaquinas(email string) ([]Maquina_virtual, error) {
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

func PowerMachine(c *gin.Context) {
	serverURL := "http://servidor_procesamiento:8081/json/startVM"

	nombre := c.PostForm("nombreMaquina")
	fmt.Println(nombre)
	clientIP := c.ClientIP()

	payload := map[string]interface{}{
		"tipo_solicitud": "start",
		"nombreVM":       nombre,
		"clientIP":       clientIP,
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

	if resp.StatusCode == http.StatusOK {

		textMessege := "¡" + mensaje + nombre + ". Por favor espere."
		c.HTML(http.StatusOK, "controlMachine.html", gin.H{
			"SuccessMessage": textMessege,
		})
	} else {
		// Registro erróneo, muestra un mensaje de error en el HTML
		textMessege := " Error al Encender " + nombre + ". Intente de nuevo."
		c.HTML(http.StatusOK, "controlMachine.html", gin.H{
			"ErrorMessage": textMessege,
		})
	}
}

func DeleteMachine(c *gin.Context) {
	serverURL := "http://servidor_procesamiento:8081/json/deleteVM"

	nombre := c.PostForm("vmnameDelete")

	payload := map[string]interface{}{
		"tipo_solicitud": "delete",
		"nombreVM":       nombre,
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

	if resp.StatusCode == http.StatusOK {
		// Registro exitoso, muestra un mensaje de éxito en el HTML
		c.HTML(http.StatusOK, "controlMachine.html", gin.H{
			"SuccessMessage": "Solicitud para eliminar la màquina enviada con èxito.",
		})
	} else {
		// Registro erróneo, muestra un mensaje de error en el HTML
		c.HTML(http.StatusOK, "controlMachine.html", gin.H{
			"ErrorMessage": "La solicitud para eliminar la màquina no fue exitosa. Intente de nuevo",
		})
	}
}

func ConfigMachine(c *gin.Context) {
	serverURL := "http://servidor_procesamiento:8081/json/modifyVM"

	// Acceder a la sesión
	session := sessions.Default(c)
	email := session.Get("email")

	if email == nil {
		// Si el usuario no está autenticado, redirige a la página de inicio de sesión
		c.Redirect(http.StatusFound, "/login")
		return
	}

	userEmail := email.(string)

	// Obtener los datos del formulario
	vmname := c.PostForm("vmnameConfig")
	memory, _ := strconv.Atoi(c.PostForm("memoryConfig"))
	cpu, _ := strconv.Atoi(c.PostForm("cpuConfig"))

	// Crear una estructura Maquina_virtual y convertirla a JSON
	Specifications := Maquina_virtual{Nombre: vmname, Ram: memory, Cpu: cpu, Persona_email: userEmail}

	fmt.Println(Specifications)

	payload := map[string]interface{}{
		"tipo_solicitud": "modify",
		"specifications": Specifications,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return
	}

	// Crear una solicitud HTTP POST con el JSON como cuerpo
	req, err := http.NewRequest("POST", serverURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return
	}

	// Establecer el encabezado de tipo de contenido
	req.Header.Set("Content-Type", "application/json")

	// Realizar la solicitud HTTP
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		// Registro exitoso, muestra un mensaje de éxito en el HTML
		c.HTML(http.StatusOK, "controlMachine.html", gin.H{
			"SuccessMessage": "Solicitud para modificar la màquina virtual enviada con èxito",
		})
	} else {
		// Registro erróneo, muestra un mensaje de error en el HTML
		c.HTML(http.StatusOK, "controlMachine.html", gin.H{
			"ErrorMessage": "La solicitud para modificar la màquina virtual no tuvo èxito. Intente de nuevo",
		})
	}
}

func GetMachines(c *gin.Context) {
	// Acceder a la sesión para obtener el email del usuario
	session := sessions.Default(c)
	email := session.Get("email")

	if email == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuario no autenticado"})
		return
	}

	userEmail := email.(string)

	// Obtener los datos de las máquinas utilizando el email del usuario
	machines, err := consultarMaquinas(userEmail)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Devolver los datos en formato JSON
	c.JSON(http.StatusOK, machines)
}

func Logout(c *gin.Context) {
	// Acceder a la sesión
	session := sessions.Default(c)
	// Eliminar la información de la sesión, incluyendo el email
	session.Delete("email")
	session.Save()

	// Redirigir al usuario a la página de inicio de sesión u otra página
	c.Redirect(http.StatusFound, "/login")
}

func EnviarContenido(c *gin.Context) {
	var data struct {
		Contenido string `json:"contenido"`
	}

	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"url": data.Contenido, // Modifica esto según tus necesidades.
	})
}

// SEGUNDA ITERACION DEKTOP CLOUD
func Checkhost(c *gin.Context) {
	session := sessions.Default(c)
	email := session.Get("email")

	if email == nil {
		// Si el usuario no está autenticado, redirige a la página de inicio de sesión
		c.Redirect(http.StatusFound, "/login")
		return
	}

	userEmail := email.(string)
	idHostStr := c.PostForm("host")
	idHost, err := strconv.Atoi(idHostStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid host ID"})
		return
	}

	fmt.Print(vmtemp)
	fmt.Print(idHost)

	// Obtener el idHost del formulario
	memoryint, _ := strconv.Atoi(vmtemp.Memory)
	cpuint, _ := strconv.Atoi(vmtemp.CPU)
	// Obtener la dirección IP del cliente
	clientIP := c.ClientIP()
	maquina_virtual := Maquina_virtual{

		Nombre:                         vmtemp.VMName,
		Sistema_operativo:              "linux",
		Distribucion_sistema_operativo: vmtemp.OS,
		Ram:                            memoryint,
		Cpu:                            cpuint,
		Persona_email:                  userEmail,
		Host_id:                        idHost}

	// Crear el objeto JSON con los datos del cliente
	payload := map[string]interface{}{
		"clientIP":       clientIP,
		"ubicacion":      idHost,
		"specifications": maquina_virtual,
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return
	}
	// Realizar una solicitud POST al servidor remoto con los datos en formato JSON
	req, err := http.NewRequest("POST", "http://servidor_procesamiento:8081/json/checkhost", bytes.NewBuffer(jsonData))
	if err != nil {
		return
	}

	// Establecer el encabezado de tipo de contenido
	req.Header.Set("Content-Type", "application/json")

	// Realizar la solicitud HTTP
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	// Verificar el código de estado de la respuesta
	if resp.StatusCode == http.StatusOK {
		c.HTML(http.StatusOK, "controlMachine.html", gin.H{"SuccessMessage": "Solicitud para chequear maquina virtual enviada con éxito."})
	} else {
		c.HTML(http.StatusOK, "controlMachine.html", gin.H{"ErrorMessage": "Esta maquina virtual tiene problemas :(  selecciona otra por favor "})
	}
}

func consultarHostDisponibles() ([]Host, error) {
	serverURL := "http://servidor_procesamiento:8081/json/consultHosts" // Cambia esto por la URL de tu servidor en el puerto 8081

	persona := Persona{Email: "123"}
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

	var hosts []Host

	// Decodifica los datos de respuesta en la variable machines.
	if err := json.Unmarshal(responseBody, &hosts); err != nil {
		// Maneja el error de decodificación aquí
	}

	return hosts, nil
}

type Maquina_virtualtemp struct {
	VMName string `json:"vmname"` // Etiquetas JSON deben coincidir
	OS     string `json:"os"`
	CPU    string `json:"cpu"` // Asegúrate de usar el tipo correcto
	Memory string `json:"memory"`
}

var vmtemp Maquina_virtualtemp

func Mvtemp(c *gin.Context) {

	// Deserializa el JSON recibido
	if err := c.ShouldBindJSON(&vmtemp); err != nil {
		fmt.Print(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Datos JSON inválidos",
		})
		return
	}
}
