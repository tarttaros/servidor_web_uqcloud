# Usa una imagen base de Go
FROM golang:1.19

# Establece el directorio de trabajo dentro del contenedor
WORKDIR /app

# Copia la carpeta desktopcloud al directorio de trabajo
COPY  Codigo .

# Compila el archivo main.go (ajusta según tus necesidades)
WORKDIR /app/servidor_web_uqcloud
RUN go build -o main main.go

# docker build -t servidor-web-compilado .  -- para crear la imagen con el codigo actual.

# docker run --name s_web -it servidor-web-compilado  --  crea el contenedor y se deja ejecutando para que se puedan extraer los ejecutables. 

# docker cp "nombre contenedor":/app/servidor_web_uqcloud/ D:\  --  se ejecuta en otra terminal para extraer la carpeta con el comprimido.