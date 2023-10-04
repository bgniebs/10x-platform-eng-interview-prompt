FROM golang:1.6-alpine

# make app directory and add contents
RUN mkdir /app 
ADD app /app/ 
WORKDIR /app 

# copy csv file
COPY seattle-weather.csv /app/

ENV BACKEND_FILENAME="seattle-weather.csv"

# go build and start
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .
CMD ["/app/main"]

EXPOSE 3000