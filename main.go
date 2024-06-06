package main

import (
	"AppWeb/handlers"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

type LoginPageData struct {
	ErrorMessage string
}

// Funcion Para la matriz de botones del caso de uso de asignacion de recursos
func mod(i, j int) int {
	return i % j
}

func main() {
	args := os.Args[1]
	port := ":" + args

	r := gin.Default()

	r.SetFuncMap(template.FuncMap{
		"json": func(v interface{}) string {
			jsonValue, _ := json.Marshal(v)
			return string(jsonValue)
		},
		"mod": mod,
	})

	// Carga las plantillas
	r.LoadHTMLGlob("templates/*.html")

	// Configurar la tienda de cookies para las sesiones
	store := cookie.NewStore([]byte("tu_clave_secreta"))
	r.Use(sessions.Sessions("sesion", store))

	// Configura las rutas
	r.LoadHTMLGlob("templates/*.html")
	r.Static("/static", "./static")

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})
	r.POST("/api/checkhost", handlers.Checkhost)
	r.GET("/login", handlers.LoginPage)
	r.GET("/signin", handlers.SigninPage)
	r.GET("/mainPage", handlers.MainPage)
	r.GET("/profile", handlers.ProfilePage)
	r.GET("/imagenes", handlers.GestionImagenes)
	r.GET("/contenedores", handlers.GestionContenedores)
	r.GET("/welcome", handlers.WelcomePage)
	r.GET("/dashboard", handlers.DashboardHandler)
	r.GET("/createHost", handlers.CreateHostPage)
	r.GET("/createDisk", handlers.CreateDiskPage)
	r.GET("/helpCenter", handlers.HelpCenterPage)
	r.GET("/aboutUs", handlers.AboutUsPage)

	r.GET("/index", handlers.Index)

	r.GET("/scrollmenu", handlers.Scrollmenu)
	r.GET("/api/machines", handlers.GetMachines)
	r.GET("/controlMachine", handlers.ControlMachine)
	r.GET("actualizaciones-maquinas", handlers.ActualizacionesMaquinas)

	r.POST("/login", handlers.Login)
	r.POST("/signin", handlers.Signin)

	r.POST("/api/createMachine", handlers.MainSend)
	r.POST("/powerMachine", handlers.PowerMachine)
	r.POST("/deleteMachine", handlers.DeleteMachine)
	r.POST("/configMachine", handlers.ConfigMachine)
	r.POST("/api/loginTemp", handlers.LoginTemp)
	r.POST("/createHost", handlers.CreateHost)
	r.POST("/createDisk", handlers.CreateDisk)
	r.POST("/DockerHub", handlers.CrearImagen)
	r.POST("/CrearImagenTar", handlers.CrearImagenArchivoTar)
	r.POST("/CrearDockerFile", handlers.CrearImagenDockerFile)
	r.POST("/eliminarImagen", handlers.EliminarImagen)
	r.POST("/eliminarImagenes", handlers.EliminarImagenes)
	r.POST("/crearContenedor", handlers.CrearContenedor)
	r.POST("/CorrerContenedor", handlers.CorrerContenedor)
	r.POST("/PausarContenedor", handlers.PausarContenedor)
	r.POST("/ReiniciarContenedor", handlers.ReiniciarContenedor)
	r.POST("/EliminarContenedor", handlers.EliminarContenedor)
	r.POST("/eliminarContenedores", handlers.EliminarContenedores)

	r.POST("/api/contendores", handlers.GetContendores)
	r.POST("/api/imagenes", handlers.GetImages)

	r.POST("/cambiar-contenido", handlers.EnviarContenido)

	r.POST("/uploadJSON", handlers.HandleUploadJSON)
	r.POST("/api/mvtemp", handlers.Mvtemp)
	r.POST("/api/checkhost", handlers.Checkhost)
	// Ruta para cerrar sesión
	r.GET("/logout", handlers.Logout)

	// Iniciar la aplicación
	err := r.Run(port)
	if err != nil {
		log.Fatal(err)
	}
}
