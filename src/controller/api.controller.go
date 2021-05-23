package controller

import (
	"bufio"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"api-sales/src/models"

	"github.com/dgraph-io/dgo/v200"
	"github.com/dgraph-io/dgo/v200/protos/api"
	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc"
)

func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Servidor a la escucha")
}

//Función para cargar los datos de las compras a la base de datos
func UploadData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var date models.Date

	//Estructura para enviar la respuesta al cliente
	type Response struct {
		Status  bool   `json:"status"`
		Message string `json:"message"`
	}

	//Lectura del request del body
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(404)
		response := Response{
			Status:  false,
			Message: "Error al leer la petición",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	err = json.Unmarshal(reqBody, &date)
	if err != nil {
		w.WriteHeader(400)
		response := Response{
			Status:  false,
			Message: "Error al leer el tipo de dato enviado",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	//Se renombra la carpeta según la fecha enviada por el cliente
	err = renameFolder(date)
	if err != nil {
		w.WriteHeader(404)
		response := Response{
			Status:  false,
			Message: "Error al cambiar el nombre de la carpeta",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	people, err := readJsonFile(date.Date)
	if err != nil {
		w.WriteHeader(404)
		response := Response{
			Status:  false,
			Message: "Error al leer los datos de los clientes",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	purchases, err := readTxtFile(date.Date)
	if err != nil {
		w.WriteHeader(404)
		response := Response{
			Status:  false,
			Message: "Error al leer los datos de las ventas",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	dg, cancel := getDgraphClient()
	defer cancel()

	//Creación de esquema de la base de datos
	ctx := context.Background()
	if err := dg.Alter(ctx, createSchema()); err != nil {
		w.WriteHeader(404)
		response := Response{
			Status:  false,
			Message: "Error al crear el esquema",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	//Generación del formato que será cargado a la base de datos
	pb, err := json.Marshal(generateCustomers(people, purchases))
	if err != nil {
		w.WriteHeader(404)
		response := Response{
			Status:  false,
			Message: "Error al hacer marshal de los consumidores",
		}
		json.NewEncoder(w).Encode(response)
		return
	}
	mu := &api.Mutation{
		CommitNow: true,
		SetJson:   pb,
	}

	//Mutación a la base de datos
	if _, err := dg.NewTxn().Mutate(ctx, mu); err != nil {
		w.WriteHeader(404)
		response := Response{
			Status:  false,
			Message: "Error al hacer la mutación de la base de datos",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	response := Response{
		Status:  true,
		Message: "Datos subidos a la base de datos con éxito",
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

//Función para obtener todos los compradores
func GetCustomers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	type AllCustomers struct {
		AllCustomers []models.Customer `json:"allCustomers"`
	}

	//Estructura que conforma la respuesta al cliente
	type Response struct {
		Status    bool         `json:"status"`
		Message   string       `json:"message"`
		Customers AllCustomers `json:"customers"`
	}

	dg, cancel := getDgraphClient()
	defer cancel()

	ctx := context.Background()

	//Estructura de la consulta que se hará a la base de datos
	query := `query AllCustomers($value: string) {
				allcustomers(func: gt(count(purchases), 0)) {
					id
					name
					age
					purchases {
						idPurchase
						ip
						device
						products {
							idProduct
							productName
							price
						}
					}
				}	  		
			}`

	//Consulta a la base de datos
	res, err := dg.NewTxn().QueryWithVars(ctx, query, map[string]string{"$value": "0"})
	if err != nil {
		w.WriteHeader(404)
		response := Response{
			Status:  false,
			Message: "Error al hacer la consulta a la base de datos",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	var allCustomers AllCustomers

	err = json.Unmarshal(res.Json, &allCustomers)
	if err != nil {
		w.WriteHeader(404)
		response := Response{
			Status:  false,
			Message: "Error en el unmarshal de la consulta",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	response := Response{
		Status:    true,
		Message:   "Consulta exitosa",
		Customers: allCustomers,
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

//Obtención del usuario por  medio de ip pasada por el cliente, determinación de compradores con la misma ip y productos más vendidos
func GetCustomer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := chi.URLParam(r, "id")

	//Estructura de la respuesta al cliente
	type Response struct {
		Status           bool                     `json:"status"`
		Message          string                   `json:"message"`
		Customer         models.Buyer             `json:"customer"`
		CustomersSameIP  []models.ResponseQueryIP `json:"customersSameIP"`
		MostSoldProducts models.SoldProduct       `json:"mostSoldProducts"`
	}

	//Consulta a la base de datos por id
	customer, err := queryCustomer(id)
	if err != nil {
		w.WriteHeader(404)
		response := Response{
			Status:  false,
			Message: "Error en la consulta a la base de datos",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	var customersIP []models.ResponseQueryIP

	//Si la longitud de los compradores es 0 significa que introdujo un id erróneo o el cliente no ha realizado compras
	if len(customer.Customer) != 0 {
		//Consumidores con la misma ip del cliente consultado
		customersWithSameIP, err := queryCustomersSameIp(id, customer.Customer)
		if err != nil {
			w.WriteHeader(404)
			response := Response{
				Status:  false,
				Message: "Error en la consulta a la base de datos",
			}
			json.NewEncoder(w).Encode(response)
			return
		}
		//Creación de lista de compradores que tienen la misma ip del comprador buscado
		customersIP = customersWithSameIP
	} else {
		w.WriteHeader(400)
		response := Response{
			Status:  false,
			Message: "Cliente inexistente o no ha realizado compras",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	//Consulta para determinar los productos más vendidos
	mostSoldProducts, err := mostSoldProducts()
	if err != nil {
		w.WriteHeader(404)
		response := Response{
			Status:  false,
			Message: "Error en la consulta a la base de datos",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	response := Response{
		Status:           true,
		Message:          "Consulta exitosa",
		Customer:         customer,
		CustomersSameIP:  customersIP,
		MostSoldProducts: mostSoldProducts,
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

//Función que permite los parámetros para la conexión con la base de datos
func getDgraphClient() (*dgo.Dgraph, models.CancelFunc) {
	conn, err := grpc.Dial("127.0.0.1:9080", grpc.WithInsecure())
	if err != nil {
		log.Fatal("While trying to dial gRPC")
	}

	dc := api.NewDgraphClient(conn)
	dg := dgo.NewDgraphClient(dc)

	return dg, func() {
		if err := conn.Close(); err != nil {
			log.Printf("Error while closing connection:%v", err)
		}
	}
}

//Esquema de la base de datos
func createSchema() *api.Operation {
	op := &api.Operation{ /*DropOp: api.Operation_ALL*/ }
	op.Schema = `
		id: string@index(exact) .
		name: string .
		age: int .
		purchases: [uid]@count@reverse .
		idPurchase: string .
		idPerson: string .
		ip: string@index(exact) .
		device: string .
		products: [uid] .
		idProduct: string .
		productName: string .
		price: string . 
	`
	return op
}

//Función para determinar los productos más vendidos
func mostSoldProducts() (_ models.SoldProduct, err error) {

	var mostSoldProducts models.SoldProduct

	dg, cancel := getDgraphClient()
	defer cancel()

	ctx := context.Background()

	//Estructura de la consulta
	variables := map[string]string{"$value": "products"}
	query := `query MostSoldProducts($value: string) {
				var(func: has(products)) @groupby(products) {
					a as count(uid)
				}
				
				mostSoldProducts(func: uid(a), orderdesc: val(a), first: 6) {
					idProduct
					productName
					total: val(a)
				}
			}`

	//Consulta a la base de datos
	res, err := dg.NewTxn().QueryWithVars(ctx, query, variables)
	if err != nil {
		return mostSoldProducts, err
	}

	err = json.Unmarshal(res.Json, &mostSoldProducts)
	if err != nil {
		return mostSoldProducts, err
	}

	return mostSoldProducts, nil
}

//Consulta del cliente por id
func queryCustomer(id string) (_ models.Buyer, err error) {
	var customer models.Buyer

	dg, cancel := getDgraphClient()
	defer cancel()

	ctx := context.Background()

	//Estructura de la consulta
	variables := map[string]string{"$id": id}
	query := `query Customer($id: string) {
				customer(func: eq(id, $id)) {
					id
					name
					age
					purchases {
						idPurchase
						idPerson
						ip
						device
						products {
							idProduct
							productName
							price
						}
					}
				}
			}`

	//Consulta a la base de datos
	res, err := dg.NewTxn().QueryWithVars(ctx, query, variables)
	if err != nil {
		return customer, err
	}

	err = json.Unmarshal(res.Json, &customer)
	if err != nil {
		return customer, err
	}

	return customer, nil
}

//Consulta para determinar cuáles clientes tienen la misma ip que el cliente buscado
func queryCustomersSameIp(id string, customer []models.Customer) (_ []models.ResponseQueryIP, err error) {
	dg, cancel := getDgraphClient()
	defer cancel()

	ctx := context.Background()

	//Estructura de la consulta
	query := `query Customers($ip: string, $id: string) {
				customers(func: eq(ip, $ip)) {
					ip
					products {
						idProduct
						productName
						price
					}
					~purchases @filter(NOT eq(id, $id)) {
						id
						name
						age
					}
				}
			}`

	var responseQueryIP models.ResponseQueryIP
	var customersWithSameIP []models.ResponseQueryIP

	//Iteración para deteminar con qué ips ha comprado el cliente
	for _, value := range customer[0].Purchases {
		variables := map[string]string{"$ip": value.Ip, "$id": id}
		res, err := dg.NewTxn().QueryWithVars(ctx, query, variables)
		if err != nil {
			return customersWithSameIP, err
		}

		err = json.Unmarshal(res.Json, &responseQueryIP)
		if err != nil {
			return customersWithSameIP, err
		}
		//Decisión para determinar si esa ip ha sido utilizada por otro cliente
		if len(responseQueryIP.Customers[0].Purchases) > 0 {
			customersWithSameIP = append(customersWithSameIP, responseQueryIP)
		}
		responseQueryIP = models.ResponseQueryIP{Customers: nil}
		variables = nil
	}

	return customersWithSameIP, err
}

//Función para renombrar la carpeta
func renameFolder(date models.Date) (err error) {
	folders, err := ioutil.ReadDir("./assets/files/")
	if err != nil {
		return err
	}

	if folders[0].Name() != date.Date {
		err = os.Rename("./assets/files/"+folders[0].Name(), "./assets/files/"+date.Date)
		if err != nil {
			return err
		}
	}

	return nil
}

//Función para generar la estructura de datos que será subida a la base de datos
func generateCustomers(people []models.Person, purchases []models.Purchase) []models.Customer {
	var purchasesPerPerson []models.Purchase
	var customers []models.Customer

	for _, person := range people {
		for _, purchase := range purchases {
			if person.Id == purchase.IdPerson {
				purchasesPerPerson = append(purchasesPerPerson, purchase)
			}
		}
		customer := models.Customer{
			Uid:       "_:" + person.Id,
			Id:        person.Id,
			Name:      person.Name,
			Age:       person.Age,
			Purchases: purchasesPerPerson,
		}
		customers = append(customers, customer)
		purchasesPerPerson = nil
	}

	return customers
}

//Función para leer el archivo json con los datos de los clientes
func readJsonFile(date string) (_ []models.Person, err error) {
	var people []models.Person

	jsonFile, err := os.Open("./assets/files/" + date + "/people.json")
	if err != nil {
		return people, err
	}
	defer jsonFile.Close()

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return people, err
	}

	err = json.Unmarshal(byteValue, &people)
	if err != nil {
		return people, err
	}
	return people, nil
}

//Función para leer archivo csv con los datos de los productos
func readCSVFile(date string) (_ []models.Product, err error) {
	var products []models.Product

	csvFile, err := os.Open("./assets/files/" + date + "/products.csv")
	if err != nil {
		return products, err
	}
	defer csvFile.Close()

	csv := csv.NewReader(bufio.NewReader(csvFile))
	csv.Comma = '\''
	csvLines, err := csv.ReadAll()
	if err != nil {
		return products, err
	}

	for _, line := range csvLines {
		product := models.Product{
			Uid:         "_:" + line[0],
			IdProduct:   line[0],
			ProductName: line[1],
			Price:       line[2],
		}
		products = append(products, product)
	}

	return products, nil
}

//Función para leer archivo txt con los datos de las compras
func readTxtFile(date string) (_ []models.Purchase, err error) {
	var purchases []models.Purchase

	data, err := ioutil.ReadFile("./assets/files/" + date + "/transactions.txt")
	if err != nil {
		return purchases, err
	}

	//Contador para determinar cuantas ventas se hicieron
	var count int = 0
	for _, char := range string(data) {
		letter := string(char)
		if letter == "#" {
			count = count + 1
		}
	}

	var word string = ""
	listSlices := make([][]string, count)
	var i int = -1
	var j int = 0

	//Ciclo para obtener la estructura buscada
	for _, numChar := range string(data) {

		char := string(numChar)
		//Si el caracter es # quiere decir que empieza una nueva compra y por eso se debe declarar un nuevo slice para esa compra
		if char == "#" {
			char = ""

			i = i + 1
			j = 0
			//Slice de 5 espacios ya que la compra tiene 5 datos
			listSlices[i] = make([]string, 5)
		}
		//Si encuentra un espacio quiere decir que terminó una palabra y esta palabra debe ser agregada al slice
		if char == " " {
			//Si la palabra está vacía quiere decir que ya va a empezar la siguiente compra, porque el archivo tiene doble espacio entre compras
			if word != "" {
				listSlices[i][j] = word
				j = j + 1
				word = ""
			}
		} else {
			word = word + char
		}
	}

	//Ciclo para darle la estructura a las compras que serán subidas a la base de datos
	for _, line := range listSlices {
		//Remueve paréntisis y le asigna el producto a cada producto de la compra
		listItems, err := assignProducts(removeBrackets(line[4]), date)
		if err != nil {
			return purchases, err
		}

		purchase := models.Purchase{
			Uid:        "_:" + line[2],
			IdPurchase: line[0],
			IdPerson:   line[1],
			Ip:         line[2],
			Device:     line[3],
			Products:   listItems,
		}
		purchases = append(purchases, purchase)
	}

	return purchases, nil
}

//Remueve los paréntesis que se tienen en el archivo txt
func removeBrackets(line string) []string {

	line = strings.ReplaceAll(line, "(", "")
	line = strings.ReplaceAll(line, ")", "")

	slice := strings.Split(line, ",")

	return slice
}

//Función para relacionar los productos de cada compra
func assignProducts(idProducts []string, date string) (_ []models.Product, err error) {
	var listItems []models.Product

	products, err := readCSVFile(date)
	if err != nil {
		return listItems, err
	}

	for i, idProduct := range idProducts {
		for j, product := range products {

			//Si el producto comprado coincide con algún producto de los registrados entonces agrega el producto con sus propiedades.
			//Si el producto comprado no coincide con ningún producto registrado entonces agréguelo sólo con id
			if idProduct == product.IdProduct {
				listItems = append(listItems, product)
			} else if j == len(products)-1 && i != len(listItems)-1 {
				listItems = append(listItems, models.Product{
					IdProduct:   idProduct,
					ProductName: "",
					Price:       "",
				})
			}
		}
	}

	return listItems, nil
}
