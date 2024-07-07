package main

import (
 "fmt"
 "os"
 "encoding/json"
 "sync"
 "github.com/jcelliott/lumber"
 "path/filepath"
"io/ioutil"
)

const Version = "1.0.0"

type(
  Logger interface {
  Fatal(string, ...interface{})
  Error(string, ...interface{})
  Warn(string, ...interface{})
  Info(string, ...interface{})
  Debug(string, ...interface{})
  Trace(string, ...interface{})
  }
  
  Driver struct {
    mutex sync.Mutex
    mutexes map[string]*sync.Mutex
    dir string
    log Logger
  }
)

type Options struct {
  Logger
}

func New(dir string, options *Options) (*Driver, error) {
  dir = filepath.Clean(dir)
  opts := Options{}
  if options != nil {
    opts = *options
  }
  if opts.Logger == nil {
    opts.Logger = lumber.NewConsoleLogger(lumber.INFO)
  }

  driver := Driver{
    dir: dir,
    mutexes: make(map[string]*sync.Mutex),
    log: opts.Logger,
  }
  if _, err := os.Stat(dir); err != nil {
    opts.Logger.Debug("using '%s'(database already exists)\n", dir)
    return &driver, nil
  }
  opts.Logger.Debug("creating '%s' database\n", dir)
  return &driver, os.MkdirAll(dir, 0755)
}
func (d *Driver) Write(collection, resource string, v interface{}) error {
if collection == "" {
    return fmt.Errorf("Missing collection")
  }
  if resource == "" {
    return fmt.Errorf("Missing resource")
  }

  mutex := d.getOrCreateMutex(collection)
  mutex.Lock()
  defer mutex.Unlock()

  dir := filepath.Join(d.dir, collection)
  fnlPath := filepath.Join(dir, resource+".json")
  tmpPath := fnlPath + ".tmp"


  if err := os.MkdirAll(dir, 0755); err != nil {
    return err
  }
  b,err := json.MarshalIndent(v, "","\t")
  if err != nil {
    return err
  }

  b = append(b, byte('\n'))

  if err := ioutil.WriteFile(tmpPath, b, 0644); err != nil {
    return err
  }
  return os.Rename(tmpPath, fnlPath)
}
func (d *Driver)Read(collection, resource string, v interface{}) error {
  if collection == "" {
    return fmt.Errorf("Missing collection")
  }
  if resource == "" {
    return fmt.Errorf("Missing resource")
  }

  record := filepath.Join(d.dir, collection, resource)
  if _, err := stat(record); err != nil {
    return err
  }
  b,err :=ioutil.ReadFile(record+".json")
  if err != nil {
    return err
  }
  return json.Unmarshal(b, &v)
}

func (d* Driver)ReadAll(collection string)([]string, error) {
  if collection == "" {
    return nil, fmt.Errorf("Missing collection")
  }
  dir := filepath.Join(d.dir, collection)
  if _, err := stat(dir); err != nil {
    return nil, err
  }
  files, _ := ioutil.ReadDir(dir)
  var records []string
  for _, file := range files {
    b, err := ioutil.ReadFile(filepath.Join(dir, file.Name()))
    if err != nil {
      return nil, err
    }
    records = append(records, string(b))
  }
  return records, nil
}

func (d * Driver)Delete(collection, resource string) error {
  path := filepath.Join( collection, resource)
  mutex := d.getOrCreateMutex(collection)
  mutex.Lock()
  defer mutex.Unlock()
  dir := filepath.Join(d.dir, path)
  switch fi, err := stat(dir); {
   case fi == nil, err != nil:
    return fmt.Errorf("Record not found %v", path)

   case fi.Mode().IsDir():
    return os.RemoveAll(dir)
  
   case fi.Mode().IsRegular():
    return os.RemoveAll(dir + ".json")
  }
  return nil
}
func (d *Driver)getOrCreateMutex( collection string) *sync.Mutex {
  d.mutex.Lock()
  defer d.mutex.Unlock()
  m, ok := d.mutexes[collection]
  if !ok {
    m = &sync.Mutex{}
    d.mutexes[collection] = m
  }
  return m
}

func stat(path string) (fi os.FileInfo, err error) {
  if fi, err = os.Stat(path); os.IsNotExist(err) {
    fi, err = os.Stat(path + ".json") 
  }
  return 
}
type Address struct {
  City string
  State string
  Country string
  Pincode json.Number
}

// create a user struct to hold the data
type User struct {
  Name string 
  Age json.Number
  Contact string
  Country string
  Address Address // Address is another struct that we will define later
}


func main() {
  dir := "./"

  db, err := New(dir,nil) //The new function in Go is used to allocate a new object of the specified type and return a pointer to it.
  if err != nil {
    fmt.Println("Error: ", err)
  }
 
  // hardcoding the user data but usually this will be from a form or an API
  employees := []User{
   {"John Doe", "25", "1234567890", "Myrl Technology", Address{"Bangalore", "Karnataka", "India", "560001"}},
   {"Jane Doe", "26", "0987654321", "Google", Address{"Mumbai", "Maharashtra", "India", "400001"}},
   {"John Smith", "27", "1234509876", "Microsoft", Address{"Los Angeles", "California", "USA", "90001"}},
   {"Jane Smith", "28", "0987612345", "Facebook", Address{"San Francisco", "California", "USA", "94016"}},
   {"John Brown", "29", "1234598760", "Remote-Teams", Address{"London", "London", "UK", "EC1A"}},
   {"Jane Brown", "30", "0987612309", "Dominate", Address{"Manchester", "Manchester", "UK", "M1"}},
  }

  for _, value := range employees {
    db.Write("users", value.Name, User{
      Name: value.Name, 
      Age: value.Age, 
      Contact: value.Contact, 
      Country: value.Country, 
      Address: value.Address,
    })
  }

  records, err := db.ReadAll("users")
  if err != nil {
    fmt.Println("Error: ", err)
  }
  fmt.Println(records)

  allusers :=[]User{}

  for _, f := range records {
    employeeFound := User{}
    if err := json.Unmarshal([]byte(f), &employeeFound); err != nil {
      fmt.Println("Error: ", err)
    }
    allusers = append(allusers, employeeFound)
  }
  fmt.Println((allusers))


  //if err := db.Delete("users", "John Doe"); err != nil {
  //fmt.Println("Errfor: ", err)}

  //if err :=db.Delete("users",""); err !=nil {
   // fmt.Println("Error: ", err)
  //}
}
