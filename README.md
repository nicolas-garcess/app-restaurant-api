### Aplicación web que permite subir datos a una base de datos local y consultarlos

## Definición del problema

Crear visualizador de clientes y transacciones de un restaurante.

### Parte 1: Cargar los datos a la base de datos

Endpoint que permite cargar los datos dada una fecha, para eso hay que procesarlos y finalmente guardarlos en la base de datos local. Por defecto cargar el día actual. Los datos a cargar son: 
- Lista de productos del día.
- lista de compradores del día. 
- Lista de transacciones del día.

### Parte 2: Listar compradores

Endpoint para listar todas las personas que han comprado en la plataforma. Debe usar la base de datos local con los datos que hayan sido cargados, el día no importa.

### Parte 3: Consultar comprador

Endpoint que recibe el ID del comprador y retorna:
- Su historial de compras.
- Otros compradores usando la misma IP.
- Algunas recomendaciones de productos que otras personas también compraron.

### Parte 4: Interfaz Web

Interfaz web con Vue que permite:
- Seleccionar una fecha para sincronizar (ver parte 1).
- Ver todos los compradores (ver parte 2).
- Ver la página del comprador con su información (ver parte 3).

## Tecnologías usadas

- Golang
- Dgraph
- Enrutador chi
- Vue-Cli y Vuetify