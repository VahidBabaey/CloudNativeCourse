package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync" // Imported to use sync.RWMutex for thread-safe operations.
)

func main() {
	// Initialize db with items and a new RWMutex for thread safety.
	db := database{
		items: map[string]dollars{"shoes": 50, "socks": 5},
		mutex: &sync.RWMutex{},
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/list", db.list)
	mux.HandleFunc("/price", db.price)
	// Register new handlers for create, update, and delete operations.
	mux.HandleFunc("/create", db.create)
	mux.HandleFunc("/update", db.update)
	mux.HandleFunc("/delete", db.delete)

	log.Fatal(http.ListenAndServe("localhost:8000", mux))
}

type dollars float32

func (d dollars) String() string { return fmt.Sprintf("$%.2f", d) } // Custom Stringer for dollars type.

// Defines a database struct with a map of items and prices, and a pointer to an RWMutex for thread safety.
type database struct {
	items map[string]dollars // Map to store item prices.
	mutex *sync.RWMutex      // Mutex to synchronize access to the items map.
}

//Handler that lists all items in the database, using a read lock for thread safety.
func (db database) list(w http.ResponseWriter, req *http.Request) {
	db.mutex.RLock() // Lock for reading to allow concurrent reads.
	defer db.mutex.RUnlock()
	for item, price := range db.items {
		fmt.Fprintf(w, "%s: %s\n", item, price)
	}
}

//Handler that shows the price of a specified item, using a read lock for thread safety.
func (db database) price(w http.ResponseWriter, req *http.Request) {
	db.mutex.RLock() // Lock for reading to allow concurrent reads.
	defer db.mutex.RUnlock()
	item := req.URL.Query().Get("item")
	if price, ok := db.items[item]; ok {
		fmt.Fprintf(w, "%s\n", price)
	} else {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "no such item: %q\n", item)
	}
}

// Handler for creating a new item in the database, using a write lock for thread safety.
func (db database) create(w http.ResponseWriter, req *http.Request) {
	db.mutex.Lock() // Lock for writing to prevent concurrent writes.
	defer db.mutex.Unlock()
	item := req.URL.Query().Get("item")
	priceStr := req.URL.Query().Get("price")
	price, err := strconv.ParseFloat(priceStr, 32) // Parse price as float32.
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "invalid price: %q\n", priceStr)
		return
	}
	if _, exists := db.items[item]; exists {
		w.WriteHeader(http.StatusConflict)
		fmt.Fprintf(w, "item already exists: %q\n", item)
	} else {
		db.items[item] = dollars(price)
		fmt.Fprintf(w, "created %s: %s\n", item, dollars(price))
	}
}

// Handler for updating the price of an existing item, using a write lock for thread safety.
func (db database) update(w http.ResponseWriter, req *http.Request) {
	db.mutex.Lock() // Lock for writing to prevent concurrent writes.
	defer db.mutex.Unlock()
	item := req.URL.Query().Get("item")
	priceStr := req.URL.Query().Get("price")
	price, err := strconv.ParseFloat(priceStr, 32) // Parse price as float32.
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "invalid price: %q\n", priceStr)
		return
	}
	if _, ok := db.items[item]; !ok {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "no such item: %q\n", item)
	} else {
		db.items[item] = dollars(price)
		fmt.Fprintf(w, "updated %s: %s\n", item, dollars(price))
	}
}

// Handler for deleting an item from the database, using a write lock for thread safety.
func (db database) delete(w http.ResponseWriter, req *http.Request) {
	db.mutex.Lock() // Lock for writing to prevent concurrent writes.
	defer db.mutex.Unlock()
	item := req.URL.Query().Get("item")
	if _, ok := db.items[item]; !ok {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "no such item: %q\n", item)
	} else {
		delete(db.items, item)
		fmt.Fprintf(w, "deleted %s\n", item)
	}
}
