# simple python tests
# TODO: use python testing module for cleaner code

import http.client
import json

connection = http.client.HTTPConnection('localhost', 3000, timeout=30)

# Test initial get of all
print("Test GET all: Start")
connection.request("GET", "/query")
response = connection.getresponse()

results = response.read().decode()

# Ensure valid json response
jresults = json.loads(results)

if response.status != 200:
    print("TEST GET all: Failed")
    raise ValueError("Invalid response code: " + response.status)
    
print("Test GET all: success")

# Test initial get of all
print("Test GET all with limit: Start")
connection.request("GET", "/query?limit=3")
response = connection.getresponse()

results = response.read().decode()

# Ensure valid json response
jresults = json.loads(results)

if response.status != 200:
    print("TEST GET all with limit: Failed")
    raise ValueError("Invalid response code: " + response.status)

# Ensure count is expected to 3
if len(jresults) != 3:
    print("TEST GET all with limit: Failed")
    raise ValueError("Unexpected size of results: " + len(jresults))

print("Test GET all with limit: success")

# Test filter by weather
print("Test GET filter by weather: Start")
connection.request("GET", "/query?weather=sun")
response = connection.getresponse()

results = response.read().decode()

# Ensure valid json response
jresults = json.loads(results)

if response.status != 200:
    print("TEST GET filter by weather: Failed")
    raise ValueError("Invalid response code: " + response.status)
    
print("Test GET filter by weather: success")

print("Test GET filter by weather and limit: Start")
connection.request("GET", "/query?weather=sun&limit=5")
response = connection.getresponse()

results = response.read().decode()

# Ensure valid json response
jresults = json.loads(results)

if response.status != 200:
    print("TEST GET filter by weather and limit: Failed")
    raise ValueError("Invalid response code: " + response.status)

# Ensure count is expected to 5
if len(jresults) != 5:
    print("TEST GET filter by weather and limit: Failed")
    raise ValueError("Unexpected size of results: " + len(jresults))

print("Test GET filter by weather and limit: success")

# Test filter by date
print("Test GET filter by date: Start")
connection.request("GET", "/query?date=2012-01-12")
response = connection.getresponse()

results = response.read().decode()

# Ensure valid json response
jresults = json.loads(results)

if response.status != 200:
    print("TEST GET filter by date: Failed")
    raise ValueError("Invalid response code: " + response.status)

if jresults[0]["date"] != "2012-01-12":
    print("TEST GET filter by date: Failed")
    raise ValueError("Invalid date return: " + jresults[0]["date"] + " expected value: " + "2012-01-12")

print("Test GET filter by date: success")

# Test filter by date
# date 2012-01-12 has sun weather, add rain filter and ensure 404
print("Test GET filter by date and weather: Start")
connection.request("GET", "/query?date=2012-01-12&weather=rain")
response = connection.getresponse()

results = response.read().decode()

# Ensure valid json response
jresults = json.loads(results)

if response.status != 404:
    print("TEST GET filter by date: Failed")
    raise ValueError("Invalid response code: " + response.status + " expected 404")

print("Test GET filter by date and weather: success")