This is a sample Go web application, allowing to
/login/logout/unregister users. Users may then add/edit/delete
named data. It tries to be as idiomatic and predictable as possible.

Both JQuery and Bootstrap are used for the UI.

# Files
## main.go
Contains all the HTTP handling. Built on top of net/http and
the gorilla toolkit, mainly:

* gorilla/schema to convert form to struct (User, Data)
* gorilla/session to store data through cookies (token)

## store.go, sqlite.go 
Store.go contains an interface for a storage system, and a
declaration of such a storage system, which is initialized in
main.go/^func main.

Sqlite.go implements such a store through SQLite.

## model.go
Is the glue between main.go and store.go: it checks data
received from user, and store or retrieve it.

# TODO

* Remove jquery/bootstrap and use CDN instead.
* Data permissions, ensure correct access
* Clean CSS/HTML, warning message, password changing, etc. (bell & whistles)
* schema: invalid path "action" on templates/index.html (inoffensive but still)
