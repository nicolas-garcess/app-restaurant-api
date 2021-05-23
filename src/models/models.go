package models

type Date struct {
	Date string `json:"date"`
}

type Product struct {
	Uid         string `json:"uid,omitempty"`
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

type CustomerIP struct {
	Ip        string    `json:"ip"`
	Products  []Product `json:"products"`
	Purchases []Person  `json:"~purchases"`
}

type ResponseQueryIP struct {
	Customers []CustomerIP `json:"customers"`
}

type Buyer struct {
	Customer []Customer `json:"customer"`
}

type Count struct {
	IdProduct   string `json:"idProduct"`
	ProductName string `json:"productName"`
	Total       int    `json:"total"`
}

type SoldProduct struct {
	Products []Count `json:"mostSoldProducts"`
}
