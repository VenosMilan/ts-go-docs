
## Index: 

### File: main.go 
 ### Package `main`

- [Structures](#structures)
- [Structure](#structure)

### File: foo.go 
 ### Package `data`

- [Foo](#foo)
- [Bar](#bar)


## Structures

#### Foo

```go
type Foo struct {
  Id                int64      `json:"id"`  
  Name              *string    `json:"name"`  
  Sub               *Bar         
  Subs              *[]Bar       
  LongNameAttribute string       
  Time              *time.Time   
}
```

#### Bar

```go
type Bar struct {
  Id  uint8   
  Yes bool    
}
```

#### Structures

```go
type Structures struct {
  Comment      string        
  StructName   string        
  StructDetail []Structure   
}
```

#### Structure

```go
type Structure struct {
  FieldName string   
  FieldType string   
  Comment   string   
  Tag       string   
}
```

