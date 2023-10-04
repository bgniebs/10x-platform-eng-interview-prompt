DEV notes:

To run:
(project-root): go run .\app\main.go

testing (curl for now):
1) curl -v -X GET "localhost:3000/query"
2) curl -v -X GET "localhost:3000/query?weather=sun&limit=5"
3) curl -v -X GET "localhost:3000/query?date=2001-02-01"

Docker instructions:

1) Build image
<project root>\: docker build .

2) Grab image id
<project root>\: docker image ls

3) run container with ports mapped to 3000
<project root>\: docker run -p 3000:3000 <image id>