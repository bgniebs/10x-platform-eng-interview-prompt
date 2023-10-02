DEV notes:

To run:
(project-root): go run .\app\main.go

testing (curl for now):
1) curl -v -X GET "localhost:3000/query"
2) curl -v -X GET "localhost:3000/query?weather=sun&limit=5"
3) curl -v -X GET "localhost:3000/query?date=2001-02-01"