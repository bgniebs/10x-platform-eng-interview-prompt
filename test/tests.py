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
    raise ValueError("Unexpected size of results: " + str(len(jresults)))

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
    raise ValueError("Unexpected size of results: " + str(len(jresults)))

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


# Test filter by date range inclusive
# 2015-12-26->2015-12-30
# expected results = 5
print("Test GET filter by date range inclusive: Start")
connection.request("GET", "/query?date>=2015-12-26&date<=2015-12-30")
response = connection.getresponse()

results = response.read().decode()

# Ensure valid json response
jresults = json.loads(results)

if response.status != 200:
    print("TEST GET filter by date: Failed")
    raise ValueError("Invalid response code: " + response.status + " expected 200")

# Ensure count is expected to 5
if len(jresults) != 5:
    print("TEST GET filter by date range inclusive: Failed")
    raise ValueError("Unexpected size of results: " + str(len(jresults)))

print("Test GET filter by date range inclusive: success")

# Test filter by date range exclusive
# 2015-12-26->2015-12-30
# expected results = 3
print("Test GET filter by date range exclusive: Start")
connection.request("GET", "/query?date>2015-12-26&date<2015-12-30")
response = connection.getresponse()

results = response.read().decode()

# Ensure valid json response
jresults = json.loads(results)

if response.status != 200:
    print("TEST GET filter by date: Failed")
    raise ValueError("Invalid response code: " + response.status + " expected 200")

# Ensure count is expected to 3
if len(jresults) != 3:
    print("TEST GET filter by date range exclusive: Failed")
    raise ValueError("Unexpected size of results: " + str(len(jresults)))

print("Test GET filter by date range exclusive: success")


# Test filter by date range no upper bound
# expected results = 6
print("Test GET filter by date range no upper bound: Start")
connection.request("GET", "/query?date>=2015-12-26")
response = connection.getresponse()

results = response.read().decode()

# Ensure valid json response
jresults = json.loads(results)

if response.status != 200:
    print("TEST GET filter by date: Failed")
    raise ValueError("Invalid response code: " + response.status + " expected 200")

# Ensure count is expected to 6
if len(jresults) != 6:
    print("TEST GET filter by date range no upper: Failed")
    raise ValueError("Unexpected size of results: " + str(len(jresults)))

print("Test GET filter by date range no upper: success")

# Test filter by date range no lower bound
# expected results = 1456
print("Test GET filter by date range no lower bound: Start")
connection.request("GET", "/query?date<=2015-12-26")
response = connection.getresponse()

results = response.read().decode()

# Ensure valid json response
jresults = json.loads(results)

if response.status != 200:
    print("TEST GET filter by date: Failed")
    raise ValueError("Invalid response code: " + response.status + " expected 200")

# Ensure count is expected to 1456
if len(jresults) != 1456:
    print("TEST GET filter by date range no lower: Failed")
    raise ValueError("Unexpected size of results: " + str(len(jresults)))

print("Test GET filter by date range no lower: success")

# Test filter by date range with limit
# expected results = 5
print("Test GET filter by date range with limit: Start")
connection.request("GET", "/query?date<=2015-12-26&limit=5")
response = connection.getresponse()

results = response.read().decode()

# Ensure valid json response
jresults = json.loads(results)

if response.status != 200:
    print("TEST GET filter by date with limit: Failed")
    raise ValueError("Invalid response code: " + response.status + " expected 200")

# Ensure count is expected to 1456
if len(jresults) != 5:
    print("TEST GET filter by date range with limit: Failed")
    raise ValueError("Unexpected size of results: " + str(len(jresults)))

print("Test GET filter by date range with limit: success")


# Test filter by date range with limit
# expected results = 638
print("Test GET filter by date range with weather: Start")
connection.request("GET", "/query?date<=2015-12-26&weather=sun")
response = connection.getresponse()

results = response.read().decode()

# Ensure valid json response
jresults = json.loads(results)

if response.status != 200:
    print("TEST GET filter by date with weather: Failed")
    raise ValueError("Invalid response code: " + response.status + " expected 200")

# Ensure count is expected to 638
if len(jresults) != 638:
    print("TEST GET filter by date range with weather: Failed")
    raise ValueError("Unexpected size of results: " + str(len(jresults)))

# TODO: iterate each value and ensure sun
print("Test GET filter by date range with weather: success")