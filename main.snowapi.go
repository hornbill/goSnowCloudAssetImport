package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func getAssetFieldVal(v SnowAssetStruct, field string) (string, bool) {
	returnVal := ""
	returnNotNil := false
	r := reflect.ValueOf(v)
	f := reflect.Indirect(r).FieldByName(field)
	if f.IsValid() && !f.IsNil() {
		returnVal = fmt.Sprintf("%v", f)
		returnNotNil = true
	}
	return returnVal, returnNotNil
}

func getCustomFieldVal(v SnowCustomFieldStruct, field string) (string, bool) {
	returnVal := ""
	returnNotNil := false
	r := reflect.ValueOf(v)
	f := reflect.Indirect(r).FieldByName(field)
	if f.IsValid() && !f.IsNil() {
		returnVal = fmt.Sprintf("%v", f)
		returnNotNil = true
	}
	return returnVal, returnNotNil
}

func getHardwareFieldVal(v SnowAssetHardwareStruct, field string) (string, bool) {
	returnVal := ""
	returnNotNil := false
	r := reflect.ValueOf(v)
	f := reflect.Indirect(r).FieldByName(field)
	if f.IsValid() && !f.IsNil() {
		returnVal = fmt.Sprintf("%v", f)
		returnNotNil = true
	}
	return returnVal, returnNotNil
}

func getNetworkFieldVal(v SnowAssetNetworkStruct, field string) (string, bool) {
	returnVal := ""
	returnNotNil := false
	r := reflect.ValueOf(v)
	f := reflect.Indirect(r).FieldByName(field)
	if f.IsValid() && !f.IsNil() {
		returnVal = fmt.Sprintf("%v", f)
		returnNotNil = true
	}
	return returnVal, returnNotNil
}

func getOpticalFieldVal(v SnowAssetOpticalDriveStruct, field string) (string, bool) {
	returnVal := ""
	returnNotNil := false
	r := reflect.ValueOf(v)
	f := reflect.Indirect(r).FieldByName(field)
	if f.IsValid() && !f.IsNil() {
		returnVal = fmt.Sprintf("%v", f)
		returnNotNil = true
	}
	return returnVal, returnNotNil
}

func getLogicalDiskFieldVal(v SnowAssetLogicalDisksStruct, field string) (string, bool) {
	returnVal := ""
	returnNotNil := false
	r := reflect.ValueOf(v)
	f := reflect.Indirect(r).FieldByName(field)
	if f.IsValid() && !f.IsNil() {
		returnVal = fmt.Sprintf("%v", f)
		returnNotNil = true
	}
	return returnVal, returnNotNil
}

func getMonitorFieldVal(v SnowAssetMonitorsStruct, field string) (string, bool) {
	returnVal := ""
	returnNotNil := false
	r := reflect.ValueOf(v)
	f := reflect.Indirect(r).FieldByName(field)
	if f.IsValid() && !f.IsNil() {
		returnVal = fmt.Sprintf("%v", f)
		returnNotNil = true
	}
	return returnVal, returnNotNil
}

func getDisplayAdapterFieldVal(v SnowAssetDisplayAdaptersStruct, field string) (string, bool) {
	returnVal := ""
	returnNotNil := false
	r := reflect.ValueOf(v)
	f := reflect.Indirect(r).FieldByName(field)
	if f.IsValid() && !f.IsNil() {
		returnVal = fmt.Sprintf("%v", f)
		returnNotNil = true
	}
	return returnVal, returnNotNil
}

func getUserFieldVal(v SnowUserDetailsStruct, field string) (string, bool) {
	returnVal := ""
	returnNotNil := false
	r := reflect.ValueOf(v)
	f := reflect.Indirect(r).FieldByName(field)
	if f.IsValid() && !f.IsNil() {
		returnVal = fmt.Sprintf("%v", f)
		returnNotNil = true
	}
	return returnVal, returnNotNil
}

func getAssetDetails(assetType AssetTypeStruct, assetID string) SnowAssetDetailsStruct {
	client := &http.Client{
		Timeout: time.Second * 30,
	}
	apiURL := APIImportConf.APIConf.APIURL + "/customers/"
	apiURL += strconv.Itoa(APIImportConf.APIConf.CustomerID) + "/"
	apiURL += assetType.RecordsAbstract + "/"
	apiURL += assetID + "/"
	apiURL += "?$format=json"
	if configDebug == true {
		logger(1, "API Request for Asset Details: "+apiURL, false)
	}

	req, err := http.NewRequest("GET", apiURL, nil)
	req.SetBasicAuth(APIImportConf.APIConf.UserName, APIImportConf.APIConf.Password)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	bodyText, err := ioutil.ReadAll(resp.Body)

	//-- New Decoder
	reader := bytes.NewReader(bodyText)
	decoder := json.NewDecoder(reader)
	//-- New Var based on APIImportConf
	decodedStruct := SnowAssetDetailsStruct{}
	//-- Decode JSON
	errDecode := decoder.Decode(&decodedStruct)
	//-- Error Checking
	if errDecode != nil {
		logger(4, "Error Decoding JSON API Response: "+fmt.Sprintf("%v", errDecode), true)
	}
	return decodedStruct
}

func getAssets(assetType AssetTypeStruct, skipRows int) SnowAssetsStruct {
	client := &http.Client{
		Timeout: time.Second * 30,
	}
	apiURL := APIImportConf.APIConf.APIURL + "/customers/"
	apiURL += strconv.Itoa(APIImportConf.APIConf.CustomerID) + "/"
	apiURL += assetType.RecordsAbstract + "/"
	apiURL += "?$inlinecount=allpages&$format=json&$orderby=" + url.QueryEscape("Id asc")
	apiURL += "&$top=" + strconv.Itoa(recordsPerPage)
	apiURL += "&$skip=" + strconv.Itoa(skipRows)
	//Add filters
	if assetType.Filter != "" {
		apiURL += "&$filter=" + url.QueryEscape(assetType.Filter)
	}
	if configDebug == true {
		logger(1, "API Request for Asset: "+apiURL, false)
	}
	req, err := http.NewRequest("GET", apiURL, nil)
	req.SetBasicAuth(APIImportConf.APIConf.UserName, APIImportConf.APIConf.Password)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	bodyText, err := ioutil.ReadAll(resp.Body)

	//-- New Decoder
	reader := bytes.NewReader(bodyText)
	decoder := json.NewDecoder(reader)
	//-- New Var based on APIImportConf
	decodedStruct := SnowAssetsStruct{}
	//-- Decode JSON
	errDecode := decoder.Decode(&decodedStruct)
	//-- Error Checking
	if errDecode != nil {
		logger(4, "Error Decoding JSON API Response: "+fmt.Sprintf("%v", errDecode), true)
	}
	if configDebug == true {
		logger(1, "Number of Assets returned by API call: "+strconv.Itoa(len(decodedStruct.Body)), false)
	}
	return decodedStruct
}

func getSnowUserField(snowField string, assetID string, userMap UserMappedFieldsStruct, assetType AssetTypeStruct) string {
	returnValue := ""
	client := &http.Client{
		Timeout: time.Second * 30,
	}
	apiURL := APIImportConf.APIConf.APIURL + "/customers/"
	apiURL += strconv.Itoa(APIImportConf.APIConf.CustomerID) + "/"
	apiURL += assetType.RecordsAbstract + "/"
	apiURL += assetID + "/"
	apiURL += assetType.UsersAbstract + "/"
	apiURL += "?$format=json&$top=1"
	//Add filters
	if userMap.Filter != "" {
		apiURL += "&$filter=" + url.QueryEscape(userMap.Filter)
	}
	//Add order
	if userMap.Order != "" {
		apiURL += "&$orderby=" + url.QueryEscape(userMap.Order)
	}
	if configDebug {
		logger(1, "API Request for User Field: "+apiURL, false)
	}

	req, err := http.NewRequest("GET", apiURL, nil)
	req.SetBasicAuth(APIImportConf.APIConf.UserName, APIImportConf.APIConf.Password)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	bodyText, err := ioutil.ReadAll(resp.Body)

	//-- New Decoder
	reader := bytes.NewReader(bodyText)
	decoder := json.NewDecoder(reader)
	//-- New Var based on APIImportConf
	decodedStruct := SnowUsersStruct{}
	//-- Decode JSON
	errDecode := decoder.Decode(&decodedStruct)
	//-- Error Checking
	if errDecode != nil {
		logger(4, "Error Decoding JSON API Response: "+fmt.Sprintf("%v", errDecode), true)
	}
	for _, userRecord := range decodedStruct.Body {
		returnValue, _ = getUserFieldVal(userRecord.User, snowField)
	}

	if userMap.StripDomainRegex != "" {
		re1, err := regexp.Compile(userMap.StripDomainRegex)
		if err != nil {
			fmt.Printf("[ERROR] %v", err)
			logger(4, "Error compiling Regular Expression: "+fmt.Sprintf("%v", err), false)
		}

		regexArr := re1.FindAllString(returnValue, 1)
		if len(regexArr) > 0 {
			returnValue = strings.Replace(returnValue, regexArr[0], "", 1)
		}
	}
	if configDebug {
		logger(1, "User Field Value returned by API call: "+returnValue, false)
	}

	return returnValue
}
