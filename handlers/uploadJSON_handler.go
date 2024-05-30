package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

func HandleUploadJSON(c *gin.Context) {
	// Obtener el archivo JSON del formulario
	file, err := c.FormFile("fileInput")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error al obtener el archivo"})
		return
	}

	// Abrir el archivo
	jsonFile, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al abrir el archivo"})
		return
	}
	defer jsonFile.Close()

	// Decodificar el archivo JSON en un mapa
	var jsonData map[string]interface{}
	err = json.NewDecoder(jsonFile).Decode(&jsonData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al leer el archivo JSON"})
		return
	}

	// Enviar los datos JSON al cliente como respuesta
	c.JSON(http.StatusOK, jsonData)
}
