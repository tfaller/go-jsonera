# go-jsonera
[![PkgGoDev](https://pkg.go.dev/badge/github.com/tfaller/go-jsonera)](https://pkg.go.dev/github.com/tfaller/go-jsonera)

jsonera is a module and cli tool to find and track changes of a json document.

## module usage
```go
// parse a json doc and init a new era document
var doc interface{}
json.Unmarshal([]byte(`{"a": 0, "b": [1, 2, 3]}`), &doc)
eraDoc := jsonera.NewEraDocument(doc)

// change the original doc and update the era doc
json.Unmarshal([]byte(`{"a": 1, "b": [1, 2]}`), &doc)
changes := eraDoc.UpdateDoc(doc)

// inspect the changes
for _, c := range changes {
    fmt.Printf("%q,%v,%v\n", jsonp.Format(c.Path), c.Era, c.Mode)
}

// optional - to track further changes marshal and save the era doc
raw, _ := json.Marshal(eraDoc)
// TODO: save raw somewhere ...
```
The printes results:
```
"/a",1,ChangeUpdate
"/b/2",1,ChangeDelete
"/b",1,ChangeUpdate
```
## cli usage
Lets have a test.json:
```js
{
    "a": 0,
    "b": [1, 2, 3]
}
```
Create an era file to track changes:
```
jsonera -json test.json -era era.json
```
Now change the original test.json:
```js
{
    "a": 1,
    "b": [1, 2]
}
```
Execute the jsonera command again. The output of the command is:
```
json-pointer,era,mode
"/a",1,ChangeUpdate
"/b/2",1,ChangeDelete
"/b",1,ChangeUpdate
```
**json-pointer** tells what property was changed.

**era** tells when the last update of this property was done.

**mode** tells what kind of change happened.
