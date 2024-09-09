module compile-main-db 

go 1.18

require (
	github.com/Tris20/FairFareFinder v0.0.1
	github.com/mattn/go-sqlite3 v1.14.22
)

replace FairFareFinder/utils/time-and-date => ../../../../../utils/time-and-date
