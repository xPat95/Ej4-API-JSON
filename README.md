# FC Barcelona Players API

REST API desarrollada en Go para gestionar jugadores del FC Barcelona utilizando un archivo JSON como almacenamiento.

La API permite consultar, crear, actualizar y eliminar jugadores utilizando distintos métodos HTTP.

---

# Tecnologías utilizadas

- Go (net/http)
- JSON como almacenamiento
- Docker
- WSL2

---

# Estructura del proyecto


go-http
│
├── main.go
├── go.mod
├── Dockerfile
├── README.md
│
└── data
└── barcelona.json


---

# Cómo ejecutar el proyecto

## Ejecutar localmente

Primero correr el servidor:


go run main.go


El servidor iniciará en:


http://localhost:24355


---

# Ejecutar con Docker

Construir la imagen:


docker build -t barca-api .


Ejecutar el contenedor:


docker run -p 24355:24355 barca-api


La API estará disponible en:


http://localhost:24355


---

# Endpoints disponibles

## Ping

Verifica que la API está funcionando.


GET /api/ping


Respuesta:


{
"message": "pong"
}


---

# Obtener todos los jugadores


GET /api/players


---

# Obtener jugador por ID


GET /api/players/{id}


Ejemplo:


GET /api/players/10


---

# Filtrar jugadores

La API permite filtrar por múltiples parámetros.

Ejemplo por nombre:


GET /api/players?name=pedri


Ejemplo por posición:


GET /api/players?position=midfielder


Ejemplo combinado:


GET /api/players?name=fer&position=midfielder


---

# Crear jugador


POST /api/players


Body JSON:


{
"full_name": "Jugador Prueba",
"shirt_number": 99,
"market_value": "1 M€",
"birth_date": "01/01/2000",
"position": "forward"
}


---

# Reemplazar jugador


PUT /api/players/{id}


---

# Actualizar parcialmente jugador


PATCH /api/players/{id}


Ejemplo:


{
"market_value": "40 M€"
}


---

# Eliminar jugador


DELETE /api/players/{id}


---

# Manejo de errores

La API devuelve errores estructurados en JSON.

Ejemplo:


{
"error": "Validation error",
"details": "shirt_number must be positive"
}


---

# Criterios implementados

✔ Métodos HTTP: GET, POST, PUT, PATCH, DELETE  
✔ Query parameters  
✔ Path parameters  
✔ Validación de datos  
✔ Manejo de errores en JSON  
✔ Persistencia en archivo JSON  
✔ Filtros combinados  
✔ Dockerfile funcional  
✔ Documentación en README

---