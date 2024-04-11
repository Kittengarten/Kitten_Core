# sqlite
A pure-go wrapper for database/sql using logoove/sqlite

## Usage

```go
type row struct {
    ID   int // pk
    Name string
}

db := &sql.Sqlite{DBPath: "demo.db"}
err := db.Open(time.Hour)
if err != nil {
    panic(err)
}

err = db.Create("demotable", &row{})
if err != nil {
    panic(err)
}

err = db.Insert("demotable", &row{ID: 123, Name: "Anna"})
if err != nil {
    panic(err)
}

var r row
err = db.Find("demotable", &r, "WHERE ID=123")
if err != nil {
    panic(err)
}
fmt.Println(r)
```
