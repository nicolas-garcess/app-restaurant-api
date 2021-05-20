package main

import (
	"api-sales/src/routes"
)

/*type task struct {
	ID      int    `json:"ID"`
	Name    string `json:"Name"`
	Content string `json:"Content"`
	Date    string `json:"Date"`
}*/

/*type Product struct {
	IdProduct   string `json:"idProduct,omitempty"`
	ProductName string `json:"productName,omitempty"`
	Price       string `json:"price,omitempty"`
}

type Purchase struct {
	Uid        string    `json:"uid,omitempty"`
	IdPurchase string    `json:"idPurchase,omitempty"`
	IdPerson   string    `json:"idPerson,omitempty"`
	Ip         string    `json:"ip,omitempty"`
	Device     string    `json:"device,omitempty"`
	Products   []Product `json:"products,omitempty"`
}

type Person struct {
	Id   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	Age  int    `json:"age,omitempty"`
}

type Customer struct {
	Uid       string     `json:"uid,omitempty"`
	Id        string     `json:"id,omitempty"`
	Name      string     `json:"name,omitempty"`
	Age       int        `json:"age,omitempty"`
	Purchases []Purchase `json:"purchases,omitempty"`
}

type CancelFunc func()

type answer struct {
	Message string `json:"message"`
}*/

func main() {
	/*router := chi.NewRouter()
	router.Get("/", index)*/
	/*router.Get("/tasks", getTasks)
	router.Post("/task", createTask)
	router.Get("/task/{id}", getTask)
	router.Delete(("/task/{id}"), deleteTask)
	router.Put(("/task/{id}"), updateTask)*/
	//router.Post("/upload-data", uploadData)

	//readCSVFile()
	//readTxtFile("1621141200")
	//uploadData()

	//log.Fatal(http.ListenAndServe(`:3000`, router))

	routes.SetUpServer("3000")
}

/*func index(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Servidor a la escucha")
}*/

/*func getDgraphClient() (*dgo.Dgraph, CancelFunc) {
	conn, err := grpc.Dial("127.0.0.1:9080", grpc.WithInsecure())
	if err != nil {
		log.Fatal("While trying to dial gRPC")
	}

	dc := api.NewDgraphClient(conn)
	dg := dgo.NewDgraphClient(dc)
	// ctx := context.Background()

	// Perform login call. If the Dgraph cluster does not have ACL and
	// enterprise features enabled, this call should be skipped.
	// for {
	// Keep retrying until we succeed or receive a non-retriable error.
	// err = dg.Login(ctx, “groot”, “password”)
	// if err == nil || !strings.Contains(err.Error(), “Please retry”) {
	// break
	// }
	// time.Sleep(time.Second)
	// }
	// if err != nil {
	// log.Fatalf(“While trying to login %v”, err.Error())
	// }

	return dg, func() {
		if err := conn.Close(); err != nil {
			log.Printf("Error while closing connection:%v", err)
		}
	}
}*/

/*type date struct {
	Date string `json:"date"`
}*/

/*func uploadData(w http.ResponseWriter, r *http.Request) {
	//var date string = "1621141200"
	var date date
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "Insert a valid date")
	}

	json.Unmarshal(reqBody, &date)

	people, err := readJsonFile(date.Date)
	if err != nil {
		fmt.Println(err)
	}
	//fmt.Println(people)
	purchases := readTxtFile(date.Date)
	//fmt.Println(purchases)

	dg, cancel := getDgraphClient()
	defer cancel()

	ctx := context.Background()
	if err := dg.Alter(ctx, createSchema()); err != nil {
		log.Fatal(err)
	}

	mu := &api.Mutation{
		CommitNow: true,
	}
	pb, err := json.Marshal(generateCustomers(people, purchases))
	if err != nil {
		log.Fatal(err)
	}

	//fmt.Println(generateCustomers(people, purchases))
	mu.SetJson = pb
	response, err := dg.NewTxn().Mutate(ctx, mu)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(response)
	answer := answer{
		Message: "Se conectó bien con " + date.Date,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(answer)
}*/

/*func generateCustomers(people []Person, purchases []Purchase) []Customer {
	var purchasesPerPerson []Purchase
	var customers []Customer

	for _, person := range people {
		for _, purchase := range purchases {
			if person.Id == purchase.IdPerson {
				purchasesPerPerson = append(purchasesPerPerson, purchase)
			}
		}
		customer := Customer{
			Uid:       "_:" + person.Id,
			Id:        person.Id,
			Name:      person.Name,
			Age:       person.Age,
			Purchases: purchasesPerPerson,
		}
		customers = append(customers, customer)
		purchasesPerPerson = nil
	}

	//fmt.Println(customers)

	return customers
}*/

/*func createSchema() *api.Operation {
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
}*/

/*func readJsonFile(date string) (_ []Person, err error) {
	jsonFile, err := os.Open("./assets/files/" + date + "/people.json")
	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var people []Person

	err = json.Unmarshal(byteValue, &people)
	if err != nil {
		//fmt.Println(err)
		return people, err
	}
	return people, nil

}*/

/*func readCSVFile(date string) (_ []Product, err error) {
	var products []Product
	csvFile, err := os.Open("./assets/files/" + date + "/products.csv")
	if err != nil {
		return products, err
	}
	defer csvFile.Close()

	csv := csv.NewReader(bufio.NewReader(csvFile))
	csv.Comma = '\''
	//csv.LazyQuotes = true
	csvLines, err := csv.ReadAll()
	if err != nil {
		return products, err
	}

	for _, line := range csvLines {
		product := Product{
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

}*/

/*func readTxtFile(date string) []Purchase {
	data, err := ioutil.ReadFile("./assets/files/" + date + "/transactions.txt")
	if err != nil {
		fmt.Println(err)
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

	var purchases []Purchase

	for _, line := range listSlices {
		listItems, err := assignProducts(removeBrackets(line[4]), date)
		if err != nil {
			fmt.Println(err)
		}

		purchase := Purchase{
			Uid:        "_:" + line[2],
			IdPurchase: line[0],
			IdPerson:   line[1],
			Ip:         line[2],
			Device:     line[3],
			Products:   listItems,
		}
		purchases = append(purchases, purchase)
	}
	//fmt.Println(purchases)
	return purchases
}*/

/*func removeBrackets(line string) []string {

	line = strings.ReplaceAll(line, "(", "")
	line = strings.ReplaceAll(line, ")", "")

	slice := strings.Split(line, ",")

	return slice
}*/

/*func assignProducts(idProducts []string, date string) (_ []Product, err error) {
	var listItems []Product
	products, err := readCSVFile(date)
	if err != nil {
		return listItems, err
	}

	for i, idProduct := range idProducts {
		for j, product := range products {

			if idProduct == product.IdProduct {
				listItems = append(listItems, product)
			} else if j == len(products)-1 && i != len(listItems)-1 {
				listItems = append(listItems, Product{
					IdProduct:   idProduct,
					ProductName: "",
					Price:       "",
				})
			}
		}
	}

	return listItems, nil
}*/
