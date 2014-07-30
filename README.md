This is a sample Go web application, allowing to
/login/logout/unregister users. Users may then add/edit/delete
named data. It tries to be as idiomatic and predictable as possible.

Both JQuery and Bootstrap are used for the UI.

# Permissions
Data can be modified only by its owner. The owner
of a Data is its creator. Data can be set public:
in this case, anyone can access it in read-only mode.

In a potential future, user should be able to create group
of users and assign properties (read/write) on their data.
Public would then be a special group. Like in a classical
filesystem.

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

* set space-time attribute for Data (gmap/openstreetmap, frequency)
* Remove jquery/bootstrap and use CDN instead.
* Data permissions, ensure correct access
* Clean CSS/HTML, warning message, password changing, etc. (bell & whistles)
* schema: invalid path "action" on templates/index.html (inoffensive but still)
* insert admin by default, add an admin panel
* a bit clumsy on ownership : to protect from user changing Uid field with handcrafted request, ensure ownership in sql; maybe cache the Data as user
