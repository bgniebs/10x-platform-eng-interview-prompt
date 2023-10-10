package main

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// Structure of an individual weather record
type WeatherRec struct {
	Date          string  `json:"date"`
	Precipitation float64 `json:"precipitation"`
	TempMax       float64 `json:"temp_max"`
	TempMin       float64 `json:"temp_min"`
	Wind          float64 `json:"wind"`
	Weather       string  `json:"weather"`
}

// Weather record list (dynamic array) - TODO move to a linked list to avoid memcopies on append
type WeatherRecList struct {
	List []*WeatherRec
}

// Structure of response payload containing pointers of records to serialize
type ResponsePayload struct {
	Results []*WeatherRec
}

// Structure containing known query parameters
type QueryParameters struct {
	Limit     int64
	Date      string
	Weather   string
	DateLower string
	DateUpper string
}

// Global record store, entire list of all records
var recordStore = WeatherRecList{List: []*WeatherRec{}}

// Global date index, hashmap of pointers to individual record with date string as key
var dateIndex = make(map[string]*WeatherRec)

// Global weather index, hashmap where key is weather and value is list of records with same weather
var weatherIndex = make(map[string]WeatherRecList)

var dateMinValue time.Time
var dateMaxValue time.Time

// Initialize the CSV file backing store
//
// Notes: Since CSV is included and will be used, assume the column format and only do row validations
//
// Log fatal errors on initialization failure, alternatively can store failed state and return 500
func initializeBackingStore() {
	// Open file, defer closing till server exits
	filename := os.Getenv("BACKEND_FILENAME")

	if filename == "" {
		log.Fatal("Enviroment variable for backend filename not found ")
	}

	f, err := os.Open(filename)

	if err != nil {
		log.Fatal("File backend open() error: ", err)
	}

	defer f.Close()

	csvReader := csv.NewReader(f)
	var lineCount int = 0
	var successCount int = 0
	var invalidCount int = 0

	for {
		rec, err := csvReader.Read()

		// Handle EOF, errors, and skip header row
		if err == io.EOF {
			break
		}

		if err != nil {
			// File corruption, panic! Alternatively, can store failed initialization state
			log.Fatal(err)
		}

		// Skip header row
		if lineCount == 0 {
			lineCount++
			continue
		}

		// Create and validate the record, add it to the record store, date index, and weather index
		//
		// TODO: since these are allocated on the HEAP need to clean up on server exit or just rely on OS to handle
		p := createWeatherRecord(rec)

		if p == nil {
			log.Println("Skipping invalid row at line: ", lineCount)
			lineCount++
			invalidCount++
			continue
		}

		populateDateMetadata(p)
		populateRecordStore(p)
		populateDateIndex(p)
		populateWeatherIndex(p)

		lineCount++
		successCount++
	}

	log.Println("Backing store initialized, rows successfully processed: ", successCount, ", invalid row count: ", invalidCount)
}

// Store record in global list store
func populateRecordStore(rec *WeatherRec) {
	recordStore.List = append(recordStore.List, rec)
}

// Update date index
func populateDateIndex(rec *WeatherRec) {
	val, exists := dateIndex[rec.Date]

	if exists {
		log.Println("Error invalid file: Duplicate date found. Record: ", val)
		return
	}

	dateIndex[rec.Date] = rec
}

// Update metadata about min and max dates
func populateDateMetadata(rec *WeatherRec) {
	// record is already valid at this point
	newDate, _ := time.Parse("2006-01-02", rec.Date)
	if dateMinValue.IsZero() || newDate.Before(dateMinValue) {
		dateMinValue = newDate
	}

	if dateMaxValue.IsZero() || newDate.After(dateMinValue) {
		dateMaxValue = newDate
	}
}

// Update the weather index list to include
func populateWeatherIndex(rec *WeatherRec) {
	indexList := weatherIndex[rec.Weather]
	indexList.List = append(weatherIndex[rec.Weather].List, rec)
	weatherIndex[rec.Weather] = indexList
}

// Validate, allocate, and populate a weather record
// date,precipitation,temp_max,temp_min,wind,weather
func createWeatherRecord(rec []string) *WeatherRec {
	precipitation, err := strconv.ParseFloat(rec[1], 64)
	if err != nil {
		log.Println("Invalid precipitation column. Error: ", err)
		return nil
	}

	tempMax, err := strconv.ParseFloat(rec[2], 64)
	if err != nil {
		log.Println("Invalid max temp column. Error: ", err)
		return nil
	}

	tempMin, err := strconv.ParseFloat(rec[3], 64)
	if err != nil {
		log.Println("Invalid min temp column. Error: ", err)
		return nil
	}

	wind, err := strconv.ParseFloat(rec[4], 64)
	if err != nil {
		log.Println("Invalid wind column. Error: ", err)
		return nil
	}

	return &WeatherRec{
		Date:          rec[0],
		Precipitation: precipitation,
		TempMax:       tempMax,
		TempMin:       tempMin,
		Wind:          wind,
		Weather:       rec[5]}
}

// weather param parsing
func parseWeatherParam(param string, qp *QueryParameters) error {
	if qp.Weather != "" {
		return errors.New("Duplicate weather parameter supplied, only 1 weather parameter allowed")
	}

	_, weatherValue, found := strings.Cut(param, "weather=")

	// Valid weather parameter found
	if found {
		qp.Weather = weatherValue
	} else {
		return errors.New("Invalid weather parameter supplied, parameter must in format weather=value")
	}

	return nil
}

// date parameter parsing state machine
func parseDateParam(param string, qp *QueryParameters) error {
	_, dateUpperValue, found := strings.Cut(param, "date<=")

	if found {
		if qp.DateUpper != "" {
			return errors.New("Duplicate date upper range parameter supplied, only 1 date upper range parameter allowed")
		}

		if qp.Date != "" {
			return errors.New("Date range parameter supplied, when single date filter parameter already supplied")
		}

		_, err := time.Parse("2006-01-02", dateUpperValue)

		if err != nil {
			return errors.New("Invalid lower bound for date parameter supplied, parameter must be in format date<=YYYY-MM-DD")
		} else {
			qp.DateUpper = dateUpperValue
			return nil
		}
	}

	_, dateLowerValue, found := strings.Cut(param, "date>=")

	if found {
		if qp.DateLower != "" {
			return errors.New("Duplicate date lower range parameter supplied, only 1 date upper range parameter allowed")
		}

		if qp.Date != "" {
			return errors.New("Date range parameter supplied, when single date filter parameter already supplied")
		}

		_, err := time.Parse("2006-01-02", dateLowerValue)

		if err != nil {
			return errors.New("Invalid lower bound for date parameter supplied, parameter must be in format date>=YYYY-MM-DD")
		} else {
			qp.DateLower = dateLowerValue
			return nil
		}
	}

	_, dateUpperValueExclusive, found := strings.Cut(param, "date<")

	if found {
		if qp.DateUpper != "" {
			return errors.New("Duplicate date upper range parameter supplied, only 1 date upper range parameter allowed")
		}

		if qp.Date != "" {
			return errors.New("Date range parameter supplied, when single date filter parameter already supplied")
		}

		dateParse, err := time.Parse("2006-01-02", dateUpperValueExclusive)

		if err != nil {
			return errors.New("Invalid upper bound for date parameter supplied, parameter must be in format date<YYYY-MM-DD")
		} else {
			// Exclusive date so subtract 1 day
			qp.DateUpper = dateParse.AddDate(0, 0, -1).Format("2006-01-02")
			return nil
		}
	}

	_, dateLowerValueExclusive, found := strings.Cut(param, "date>")

	if found {
		if qp.DateLower != "" {
			return errors.New("Duplicate date lower range parameter supplied, only 1 date upper range parameter allowed")
		}

		if qp.Date != "" {
			return errors.New("Date range parameter supplied, when single date filter parameter already supplied")
		}

		dateParse, err := time.Parse("2006-01-02", dateLowerValueExclusive)

		if err != nil {
			return errors.New("Invalid lower bound for date parameter supplied, parameter must be in format date>YYYY-MM-DD")
		} else {
			// Exclusive range so add 1 day
			qp.DateLower = dateParse.AddDate(0, 0, 1).Format("2006-01-02")
			return nil
		}
	}

	_, dateSingle, found := strings.Cut(param, "date=")

	if found {
		if qp.Date != "" {
			return errors.New("Duplicate date parameter supplied, only 1 date parameter allowed")
		}

		if qp.DateLower != "" || qp.DateUpper != "" {
			return errors.New("Single date constraint supplied, when range date parameters already supplied")
		}

		_, err := time.Parse("2006-01-02", dateSingle)

		if err != nil {
			return errors.New("Invalid date parameter supplied, parameter must be in format date=YYYY-MM-DD")
		} else {
			qp.Date = dateSingle
			return nil
		}
	}

	return errors.New("Invalid date parameter, valid operators for date parameter are =, <=, <, >=, >")
}

// Limit parameter parsing
func parseLimitParam(param string, qp *QueryParameters) error {
	_, limitValue, found := strings.Cut(param, "limit=")

	if found {
		if qp.Limit != -1 {
			return errors.New("Duplicate limit parameter supplied, only 1 limit parameter allowed")
		}

		limit, err := strconv.ParseInt(limitValue, 0, 64)

		if err != nil || limit <= 0 {
			return errors.New("Invalid limit parameter supplied, parameter must be a postive and non-zero integer")
		} else {
			qp.Limit = limit
		}
	} else {
		return errors.New("Invalid limit parameter supplied, parameter must in format limit=value")
	}

	return nil
}

// Get the well-known query string parameters and validate them
func getQueryStringParameters(r *http.Request) (QueryParameters, error) {
	var qp = QueryParameters{Limit: -1, Date: "", Weather: "", DateLower: "", DateUpper: ""}
	querystring := r.URL.RawQuery

	params := strings.Split(querystring, "&")

	for _, param := range params {
		// weather parameter
		if strings.Contains(param, "weather") {
			err := parseWeatherParam(param, &qp)

			if err != nil {
				return qp, err
			}
		}

		if strings.Contains(param, "date") {
			err := parseDateParam(param, &qp)

			if err != nil {
				return qp, err
			}
		}

		if strings.Contains(param, "limit") {
			err := parseLimitParam(param, &qp)

			if err != nil {
				return qp, err
			}
		}
	}

	log.Println("Query Paremeters: Date: ", qp.Date, " DateLower: ", qp.DateLower, " DateUpper: ", qp.DateUpper, " Weather: ", qp.Weather, " Limit: ", qp.Limit)

	return qp, nil
}

// Wrapper to check if a parameter string exists
func parameterExists(s *string) bool {
	var result = false

	if len(*s) > 0 {
		result = true
	}

	return result
}

// Write HTTP respond headers and content type
func writeHeaders(w *http.ResponseWriter, statusCode int) {
	(*w).Header().Set("Content-Type", "application/json")
	(*w).WriteHeader(statusCode)
}

// Serialize result records into json and write the response
func writeJsonPayload(w *http.ResponseWriter, payload *ResponsePayload) {
	jsonResp, err := json.Marshal(payload.Results)

	if err != nil {
		log.Println("Failed to convert response payload to json. Error: ", err)
	}

	_, err = fmt.Fprintf(*w, string(jsonResp))

	if err != nil {
		log.Println("Failed to write json to response writer. Error: ", err)
	}
}

// Serialize error message into json ("message": "string of message) and write response
func writeErrorResponseMessage(w *http.ResponseWriter, errMessage string) {
	resp := make(map[string]string)
	resp["message"] = errMessage

	jsonResp, err := json.Marshal(resp)

	if err != nil {
		log.Println("Failed to convert response payload to json. Error: ", err)
	}

	_, err = fmt.Fprintf(*w, string(jsonResp))

	if err != nil {
		log.Println("Failed to write json to response writer. Error: ", err)
	}
}

// Filter records by date using the date filter supplied
// Notes: limit is not needed to be applied here since hard constraint of 1 date, but ensure weather filter is considered as well
func getDateFilter(resp *ResponsePayload, qp *QueryParameters) {
	val, exists := dateIndex[qp.Date]
	var processedrecordStore = 0

	if exists {
		if parameterExists(&qp.Weather) {
			if val.Weather == qp.Weather {
				resp.Results = append(resp.Results, val)
				processedrecordStore++
			}
		} else {
			resp.Results = append(resp.Results, val)
			processedrecordStore++
		}
	}

	log.Println("Date filter applied, record(s) processed: ", processedrecordStore)
}

// Filter records by date range using the date filter supplied
func getDateRangeFilter(resp *ResponsePayload, qp *QueryParameters) {
	var curDate = dateMinValue
	var maxDate = dateMaxValue
	var limitExists = qp.Limit > -1
	var processedrecordStore int64 = 0
	var processingComplete = false

	if qp.DateLower != "" {
		lowerBound, _ := time.Parse("2006-01-02", qp.DateLower)

		// If the lower bound is after the min date then apply lower bound filter
		if curDate.Before(lowerBound) {
			curDate = lowerBound
		}
	}

	if qp.DateUpper != "" {
		upperBound, _ := time.Parse("2006-01-02", qp.DateUpper)

		// If the max bound is after the max date then apply max bound filter
		if maxDate.After(upperBound) {
			maxDate = upperBound
		}
	}

	for !processingComplete {
		// Break if limit has been applied
		if limitExists && processedrecordStore >= qp.Limit {
			processingComplete = true
			continue
		}

		// Break if range is complete
		if curDate.After(maxDate) {
			processingComplete = true
			continue
		}

		val, exists := dateIndex[curDate.Format("2006-01-02")]

		if exists {
			if parameterExists(&qp.Weather) {
				if val.Weather == qp.Weather {
					resp.Results = append(resp.Results, val)
					processedrecordStore++
				}
			} else {
				resp.Results = append(resp.Results, val)
				processedrecordStore++
			}
		}

		curDate = curDate.AddDate(0, 0, 1)
	}

	log.Println("Date range filter applied, record(s) processed: ", processedrecordStore)
}

// Filter by weather tag, apply limit is present in parameters
func getWeatherFilter(resp *ResponsePayload, qp *QueryParameters) {
	var limitExists = qp.Limit > -1
	var processedrecordStore int64 = 0
	val, _ := weatherIndex[qp.Weather]

	for _, item := range val.List {
		if limitExists && processedrecordStore >= qp.Limit {
			break
		}

		resp.Results = append(resp.Results, item)
		processedrecordStore++
	}

	log.Println("Weather filter applied, record(s) processed: ", processedrecordStore)
}

// Retrieve data from global list in order, and enforce limit if applied
func getData(resp *ResponsePayload, qp *QueryParameters) {
	var limitExists = qp.Limit > -1
	var processedrecordStore int64 = 0

	for _, item := range recordStore.List {
		if limitExists && processedrecordStore >= qp.Limit {
			break
		}

		resp.Results = append(resp.Results, item)
		processedrecordStore++
	}

	log.Println("Data retrieved, record(s) processed: ", processedrecordStore)
}

// HTTP handler for /query
func handleGet(w http.ResponseWriter, r *http.Request) {
	var qp, err = getQueryStringParameters(r)

	// Error out if query parameters were invalid
	if err != nil {
		writeHeaders(&w, http.StatusBadRequest)
		writeErrorResponseMessage(&w, err.Error())
		return
	}

	var response = ResponsePayload{Results: []*WeatherRec{}}

	if parameterExists(&qp.Date) {
		getDateFilter(&response, &qp)
	} else if parameterExists(&qp.DateLower) || parameterExists(&qp.DateUpper) {
		getDateRangeFilter(&response, &qp)
	} else if parameterExists(&qp.Weather) {
		getWeatherFilter(&response, &qp)
	} else {
		getData(&response, &qp)
	}

	// If no results found set status to 404 and add message
	if len(response.Results) <= 0 {
		writeHeaders(&w, http.StatusNotFound)
		writeErrorResponseMessage(&w, "No results found")
		return
	}

	// Serialize to json and write response
	writeHeaders(&w, http.StatusOK)
	writeJsonPayload(&w, &response)
}

// main function
func main() {
	// Initialize CSV file backing store
	initializeBackingStore()

	// Register route
	http.HandleFunc("/query", handleGet)

	// Listen
	http.ListenAndServe(":3000", nil)
}
