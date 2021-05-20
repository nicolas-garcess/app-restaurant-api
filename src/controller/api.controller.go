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

func UploadData(w http.ResponseWriter, r *http.Request) {
	//var date string = "1621141200"
	var date models.Date
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	err = json.Unmarshal(reqBody, &date)
	if err != nil {
		http.Error(w, err.Error(), 404)
		return
	}

	people, err := readJsonFile(date.Date)
	if err != nil {
		http.Error(w, err.Error(), 404)
		return
	}

	purchases, err := readTxtFile(date.Date)
	if err != nil {
		http.Error(w, err.Error(), 404)
		return
	}

	dg, cancel := getDgraphClient()
	defer cancel()

	ctx := context.Background()
	if err := dg.Alter(ctx, createSchema()); err != nil {
		http.Error(w, err.Error(), 404)
		return
	}

	/*mu := &api.Mutation{
		CommitNow: true,
	}*/
	pb, err := json.Marshal(generateCustomers(people, purchases))
	if err != nil {
		http.Error(w, err.Error(), 404)
	}
	mu := &api.Mutation{
		CommitNow: true,
		SetJson:   pb,
	}

	//mu.SetJson = pb
	if _, err := dg.NewTxn().Mutate(ctx, mu); err != nil {
		http.Error(w, err.Error(), 404)
		return
	}

	answer := models.Answer{
		Message: "Se conectó bien con " + date.Date,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(answer)
}

func GetCustomers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	dg, cancel := getDgraphClient()
	defer cancel()

	ctx := context.Background()

	query := `query all($a: string) {
				customers(func: gt(count(purchases), 0)) {
					id
					name
					age
					purchases {
						idPurchase
						ip
						device
						products {
							idProduct
							nameProduct
							price
						}
					}
				}	  		
			}`

	res, err := dg.NewTxn().QueryWithVars(ctx, query, map[string]string{"$a": "0"})
	if err != nil {
		http.Error(w, err.Error(), 404)
		return
	}

	json.NewEncoder(w).Encode(string(res.Json))
}

func GetCustomer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := chi.URLParam(r, "id")

	dg, cancel := getDgraphClient()
	defer cancel()

	ctx := context.Background()

	varib, queryCustomer := queryCustomer(id)

	res, err := dg.NewTxn().QueryWithVars(ctx, queryCustomer, varib)
	if err != nil {
		http.Error(w, err.Error(), 404)
		return
	}

	type Buyer struct {
		Customer []models.Customer `json:"customer"`
	}

	var customer Buyer
	err = json.Unmarshal(res.Json, &customer)
	if err != nil {
		http.Error(w, err.Error(), 404)
		return
	}

	type CustomerIP struct {
		Ip        string           `json:"ip"`
		Products  []models.Product `json:"products"`
		Purchases []models.Person  `json:"~purchases"`
	}

	type SameIP struct {
		Customers []CustomerIP `json:"customers"`
	}

	var customerSameIP SameIP
	var customersWithSameIP []SameIP

	if len(customer.Customer) != 0 {
		//Iteración para deteminar con qué ips ha comprado el cliente
		for _, value := range customer.Customer[0].Purchases {
			variables, queryIP := queryIp(value.Ip, id)
			res, err := dg.NewTxn().QueryWithVars(ctx, queryIP, variables)
			if err != nil {
				http.Error(w, err.Error(), 404)
				return
			}

			err = json.Unmarshal(res.Json, &customerSameIP)
			if err != nil {
				http.Error(w, err.Error(), 404)
				return
			}

			if len(customerSameIP.Customers[0].Purchases) > 0 {
				customersWithSameIP = append(customersWithSameIP, customerSameIP)
			}
			customerSameIP = SameIP{Customers: nil}
		}
	} else {
		http.Error(w, "No existe usuario con el ID buscado", 400)
		return
	}

	type Reply struct {
		Customer        Buyer    `json:"customer"`
		CustomersSameIP []SameIP `json:"customersSameIP"`
	}
	reply := Reply{
		Customer:        customer,
		CustomersSameIP: customersWithSameIP,
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(reply)
}

func queryCustomer(id string) (map[string]string, string) {
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

	return variables, query
}

func queryIp(ip string, id string) (map[string]string, string) {
	variables := map[string]string{"$ip": ip, "$id": id}
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

	return variables, query
}

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

func createSchema() *api.Operation {
	op := &api.Operation{}
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
			IdProduct:   line[0],
			ProductName: line[1],
			Price:       line[2],
		}
		products = append(products, product)
	}

	return products, nil

	/*date := time.Date(2021, time.May, 16, 10, 51, 0, 0, time.Local)
	unix := date.Unix()
	dateUnix := Date{
		DateUnix: unix,
	}
	fmt.Println(dateUnix)*/
}

func readTxtFile(date string) (_ []models.Purchase, err error) {
	var purchases []models.Purchase
	data, err := ioutil.ReadFile("./assets/files/" + date + "/transactions.txt")
	if err != nil {
		return purchases, err
	}

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

	//For para obtener la estructura buscada
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

	for _, line := range listSlices {
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

func removeBrackets(line string) []string {

	line = strings.ReplaceAll(line, "(", "")
	line = strings.ReplaceAll(line, ")", "")

	slice := strings.Split(line, ",")

	return slice
}

func assignProducts(idProducts []string, date string) (_ []models.Product, err error) {
	var listItems []models.Product

	products, err := readCSVFile(date)
	if err != nil {
		return listItems, err
	}

	for i, idProduct := range idProducts {
		for j, product := range products {

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
