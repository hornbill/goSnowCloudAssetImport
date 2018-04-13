package main

import (
	"encoding/xml"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/hornbill/goApiLib"
	"github.com/hornbill/pb"
)

//getAssetClass -- Get Asset Class & Type ID from Asset Type Name

func getAssetClass(confAssetType string) (assetClass string, assetType int) {
	espXmlmc := apiLib.NewXmlmcInstance(APIImportConf.URL)
	espXmlmc.SetAPIKey(APIImportConf.APIKey)
	espXmlmc.SetParam("application", appServiceManager)
	espXmlmc.SetParam("entity", "AssetsTypes")
	espXmlmc.OpenElement("searchFilter")
	espXmlmc.SetParam("h_name", confAssetType)
	espXmlmc.CloseElement("searchFilter")
	espXmlmc.SetParam("maxResults", "1")
	var XMLSTRING = espXmlmc.GetParam()
	XMLGetMeta, xmlmcErr := espXmlmc.Invoke("data", "entityBrowseRecords")
	if xmlmcErr != nil {
		logger(4, "API Call failed when retrieving Asset Class:"+fmt.Sprintf("%v", xmlmcErr), false)
		logger(1, "API XML: "+fmt.Sprintf("%s", XMLSTRING), false)
	}

	var xmlRespon xmlmcTypeListResponse
	err := xml.Unmarshal([]byte(XMLGetMeta), &xmlRespon)
	if err != nil {
		logger(4, "Could not get Asset Class and Type. Please check AssetType within your configuration file:"+fmt.Sprintf("%v", err), true)
		logger(1, "API XML: "+fmt.Sprintf("%s", XMLSTRING), false)
	} else {
		assetClass = xmlRespon.Params.RowData.Row.TypeClass
		assetType = xmlRespon.Params.RowData.Row.TypeID
	}
	return
}

//processAssets -- Processes Assets from Asset Map
//--If asset already exists on the instance, update
//--If asset doesn't exist, create
func processAssets(arrAssets []SnowAssetBodyStruct, assetType AssetTypeStruct) {

	logger(1, "Processing batch of "+strconv.Itoa(len(arrAssets))+" assets", true)

	bar := pb.StartNew(len(arrAssets))
	//Get the identity of the AssetID field from the config
	assetIDIdent := fmt.Sprintf("%v", assetType.AssetID)

	//-- Loop each asset
	maxGoroutinesGuard := make(chan struct{}, maxGoroutines)

	for _, assetRecord := range arrAssets {
		maxGoroutinesGuard <- struct{}{}
		worker.Add(1)
		assetMap := assetRecord.Asset

		go func() {
			defer worker.Done()
			mutexBar.Lock()
			bar.Increment()
			mutexBar.Unlock()

			arrAssetIDIdent := strings.Split(assetIDIdent, ".")
			assetID, assetValReturned := getAssetFieldVal(assetMap, arrAssetIDIdent[1])
			assetSnowPk, _ := getAssetFieldVal(assetMap, "ID")
			//Get detailed asset info
			assetDetails := getAssetDetails(assetType, assetSnowPk)
			assetMap = assetDetails.Asset
			if assetValReturned == true {
				espXmlmc := apiLib.NewXmlmcInstance(APIImportConf.URL)
				espXmlmc.SetAPIKey(APIImportConf.APIKey)

				var boolUpdate = false
				boolUpdate, assetIDInstance := getAssetID(assetID, espXmlmc)
				//-- Update or Create Asset
				if boolUpdate && APIImportConf.UpdateAssets {
					logger(1, "Update Asset: "+assetID+" ["+assetIDInstance+"]", false)
					updateAsset(assetMap, assetIDInstance, espXmlmc, assetType)
				}
				if !boolUpdate && APIImportConf.AddAssets {
					logger(1, "Create Asset: "+assetID, false)
					createAsset(assetMap, espXmlmc, assetType)
				}
			}

			<-maxGoroutinesGuard
		}()
	}
	worker.Wait()
	bar.FinishPrint("Processing Complete!")
}

//getAssetID -- Check if asset is on the instance
//-- Returns true, assetid if so
//-- Returns false, "" if not
func getAssetID(assetName string, espXmlmc *apiLib.XmlmcInstStruct) (bool, string) {
	boolReturn := false
	returnAssetID := ""
	espXmlmc.SetParam("application", appServiceManager)
	espXmlmc.SetParam("entity", "Asset")
	espXmlmc.OpenElement("searchFilter")
	espXmlmc.SetParam("h_name", assetName)
	espXmlmc.CloseElement("searchFilter")
	espXmlmc.SetParam("maxResults", "1")
	var XMLSTRING = espXmlmc.GetParam()
	XMLAssetSearch, xmlmcErr := espXmlmc.Invoke("data", "entityBrowseRecords")
	if xmlmcErr != nil {
		logger(4, "API Call failed when searching instance for existing Asset:"+fmt.Sprintf("%v", xmlmcErr), false)
		logger(1, "API Call XML: "+fmt.Sprintf("%s", XMLSTRING), false)
	} else {
		var xmlRespon xmlmcAssetResponse
		err := xml.Unmarshal([]byte(XMLAssetSearch), &xmlRespon)
		if err != nil {
			logger(4, "Unable to Search for Asset: "+fmt.Sprintf("%v", err), true)
			logger(1, "API Call XML: "+fmt.Sprintf("%s", XMLSTRING), false)
		} else {
			if xmlRespon.MethodResult != "ok" {
				logger(4, "Unable to Search for Asset: "+xmlRespon.State.ErrorRet, true)
				logger(1, "API Call XML: "+fmt.Sprintf("%s", XMLSTRING), false)
			} else {
				returnAssetID = xmlRespon.Params.RowData.Row.AssetID
				//-- Check Response
				if returnAssetID != "" {
					boolReturn = true
				}
			}
		}
	}
	return boolReturn, returnAssetID
}

// createAsset -- Creates Asset record from the passed through map data
func createAsset(u SnowAssetStruct, espXmlmc *apiLib.XmlmcInstStruct, assetType AssetTypeStruct) {
	//Get site ID
	siteID := ""
	siteNameMapping := fmt.Sprintf("%v", APIImportConf.AssetGenericFieldMapping["h_site"])
	siteName := getFieldValue("h_site", siteNameMapping, u, assetType)
	if siteName != "" {
		siteIsInCache, SiteIDCache := siteInCache(siteName)
		//-- Check if we have cached the site already
		if siteIsInCache {
			siteID = strconv.Itoa(SiteIDCache)
		} else {
			siteIsOnInstance, SiteIDInstance := searchSite(siteName, espXmlmc)
			//-- If Returned set output
			if siteIsOnInstance {
				siteID = strconv.Itoa(SiteIDInstance)
			}
		}
	}

	//Get Owned By name
	ownedByName := ""
	ownedByURN := ""
	ownedByMapping := fmt.Sprintf("%v", APIImportConf.AssetGenericFieldMapping["h_owned_by"])
	ownedByID := getFieldValue("h_owned_by", ownedByMapping, u, assetType)
	if ownedByID != "" {
		ownedByIsInCache, ownedByNameCache := customerInCache(ownedByID)
		//-- Check if we have cached the customer already
		if ownedByIsInCache {
			ownedByName = ownedByNameCache
		} else {
			ownedByIsOnInstance, ownedByNameInstance := searchCustomer(ownedByID, espXmlmc)
			//-- If Returned set output
			if ownedByIsOnInstance {
				ownedByName = ownedByNameInstance
			}
		}
	}
	if ownedByName != "" {
		ownedByURN = "urn:sys:0:" + ownedByName + ":" + ownedByID
	}

	lastLoggedOnByURN := ""
	lastLoggedOnUserMapping := fmt.Sprintf("%v", APIImportConf.AssetTypeFieldMapping["h_last_logged_on_user"])
	if lastLoggedOnUserMapping != "" {
		lastLoggedOnByID := getFieldValue("h_last_logged_on_user", lastLoggedOnUserMapping, u, assetType)
		if lastLoggedOnByID != "" {
			lastLoggedOnByIsInCache, lastLoggedOnByNameCache := customerInCache(lastLoggedOnByID)
			//-- Check if we have cached the customer already
			if lastLoggedOnByIsInCache {
				lastLoggedOnByURN = "urn:sys:0:" + lastLoggedOnByNameCache + ":" + lastLoggedOnByID
			} else {
				lastLoggedOnByIsOnInstance, lastLoggedOnByNameInstance := searchCustomer(lastLoggedOnByID, espXmlmc)
				//-- If Returned set output
				if lastLoggedOnByIsOnInstance {
					lastLoggedOnByURN = "urn:sys:0:" + lastLoggedOnByNameInstance + ":" + lastLoggedOnByID
				}
			}
		}
	}

	//Get Used By name
	usedByName := ""
	usedByURN := ""
	usedByMapping := fmt.Sprintf("%v", APIImportConf.AssetGenericFieldMapping["h_used_by"])
	usedByID := getFieldValue("h_used_by", usedByMapping, u, assetType)
	if usedByID != "" {
		usedByIsInCache, usedByNameCache := customerInCache(usedByID)
		//-- Check if we have cached the customer already
		if usedByIsInCache {
			usedByName = usedByNameCache
		} else {
			usedByIsOnInstance, usedByNameInstance := searchCustomer(usedByID, espXmlmc)
			//-- If Returned set output
			if usedByIsOnInstance {
				usedByName = usedByNameInstance
			}
		}
	}
	if usedByName != "" {
		usedByURN = "urn:sys:0:" + usedByName + ":" + usedByID
	}

	//Get/Set params from map stored against FieldMapping
	strAttribute := ""
	strMapping := ""
	espXmlmc.SetParam("application", appServiceManager)
	espXmlmc.SetParam("entity", "Asset")
	espXmlmc.SetParam("returnModifiedData", "true")
	espXmlmc.OpenElement("primaryEntityData")
	espXmlmc.OpenElement("record")
	//Set Class & TypeID
	espXmlmc.SetParam("h_class", AssetClass)
	espXmlmc.SetParam("h_type", strconv.Itoa(AssetTypeID))

	espXmlmc.SetParam("h_last_updated", APITimeNow)
	espXmlmc.SetParam("h_last_updated_by", "Import - Add")

	//Get asset field mapping
	for k, v := range APIImportConf.AssetGenericFieldMapping {
		strAttribute = fmt.Sprintf("%v", k)
		strMapping = fmt.Sprintf("%v", v)
		if strAttribute == "h_used_by" && usedByName != "" && usedByURN != "" {
			espXmlmc.SetParam("h_used_by", usedByURN)
			espXmlmc.SetParam("h_used_by_name", usedByName)
		}
		if strAttribute == "h_owned_by" && ownedByName != "" && ownedByURN != "" {
			espXmlmc.SetParam("h_owned_by", ownedByURN)
			espXmlmc.SetParam("h_owned_by_name", ownedByName)
		}
		if strAttribute == "h_site" && siteID != "" && siteName != "" {
			espXmlmc.SetParam("h_site", siteName)
			espXmlmc.SetParam("h_site_id", siteID)
		}
		if strAttribute != "h_site" &&
			strAttribute != "h_used_by" &&
			strAttribute != "h_owned_by" &&
			strMapping != "" && getFieldValue(strAttribute, strMapping, u, assetType) != "" {
			espXmlmc.SetParam(strAttribute, getFieldValue(strAttribute, strMapping, u, assetType))
		}
	}
	espXmlmc.CloseElement("record")
	espXmlmc.CloseElement("primaryEntityData")

	//Add extended asset type field mapping
	espXmlmc.OpenElement("relatedEntityData")
	//Set Class & TypeID
	espXmlmc.SetParam("relationshipName", "AssetClass")
	espXmlmc.SetParam("entityAction", "insert")
	espXmlmc.OpenElement("record")
	espXmlmc.SetParam("h_type", strconv.Itoa(AssetTypeID))
	//Get asset field mapping
	for k, v := range APIImportConf.AssetTypeFieldMapping {
		strAttribute = fmt.Sprintf("%v", k)
		strMapping = fmt.Sprintf("%v", v)
		if strAttribute == "h_last_logged_on_user" && lastLoggedOnByURN != "" {
			espXmlmc.SetParam("h_last_logged_on_user", lastLoggedOnByURN)
		}
		if strAttribute != "h_last_logged_on_user" &&
			strMapping != "" && getFieldValue(strAttribute, strMapping, u, assetType) != "" {
			espXmlmc.SetParam(strAttribute, getFieldValue(strAttribute, strMapping, u, assetType))
		}
	}
	espXmlmc.CloseElement("record")
	espXmlmc.CloseElement("relatedEntityData")
	var XMLSTRING = espXmlmc.GetParam()
	//-- Check for Dry Run
	if configDryRun != true {
		if configDebug {
			logger(1, "Asset Insert XML "+fmt.Sprintf("%s", XMLSTRING), false)
		}
		XMLCreate, xmlmcErr := espXmlmc.Invoke("data", "entityAddRecord")
		if xmlmcErr != nil {
			logger(4, "Error running entityAddRecord API for createAsset:"+fmt.Sprintf("%v", xmlmcErr), false)
			logger(1, "API Call XML: "+fmt.Sprintf("%s", XMLSTRING), false)
			return
		}
		var xmlRespon xmlmcUpdateResponse

		err := xml.Unmarshal([]byte(XMLCreate), &xmlRespon)
		if err != nil {
			mutexCounters.Lock()
			counters.createskipped++
			mutexCounters.Unlock()
			logger(4, "Unable to read response from Hornbill instance from entityAddRecord API for createAsset:"+fmt.Sprintf("%v", err), false)
			logger(1, "API Call XML: "+fmt.Sprintf("%s", XMLSTRING), false)
			return
		}
		if xmlRespon.MethodResult != "ok" {
			logger(4, "Unable to add asset: "+xmlRespon.State.ErrorRet, false)
			logger(1, "API Call XML: "+fmt.Sprintf("%s", XMLSTRING), false)
			mutexCounters.Lock()
			counters.createskipped++
			mutexCounters.Unlock()
		} else {
			mutexCounters.Lock()
			counters.created++
			mutexCounters.Unlock()
			assetID := xmlRespon.UpdatedCols.AssetPK
			//Now add asset URN
			espXmlmc.SetParam("application", "com.hornbill.servicemanager")
			espXmlmc.SetParam("entity", "Asset")
			espXmlmc.OpenElement("primaryEntityData")
			espXmlmc.OpenElement("record")
			espXmlmc.SetParam("h_pk_asset_id", assetID)
			espXmlmc.SetParam("h_asset_urn", "urn:sys:entity:com.hornbill.servicemanager:Asset:"+assetID)
			espXmlmc.CloseElement("record")
			espXmlmc.CloseElement("primaryEntityData")
			XMLSTRING = espXmlmc.GetParam()
			XMLUpdate, xmlmcErr := espXmlmc.Invoke("data", "entityUpdateRecord")
			if xmlmcErr != nil {
				logger(4, "API Call failed when Updating Asset URN:"+fmt.Sprintf("%v", xmlmcErr), false)
				return
			}
			var xmlRespon xmlmcResponse

			err := xml.Unmarshal([]byte(XMLUpdate), &xmlRespon)
			if err != nil {
				logger(4, "Unable to read response from Hornbill instance when Updating Asset URN:"+fmt.Sprintf("%v", err), false)
				return
			}
			if xmlRespon.MethodResult != "ok" {
				logger(4, "Unable to update Asset URN: "+xmlRespon.State.ErrorRet, false)
				logger(1, "API Call XML: "+fmt.Sprintf("%s", XMLSTRING), false)
				return
			}
			return
		}
	} else {
		//-- DEBUG XML TO LOG FILE
		logger(1, "Asset Create XML "+fmt.Sprintf("%s", XMLSTRING), false)
		mutexCounters.Lock()
		counters.createskipped++
		mutexCounters.Unlock()
		espXmlmc.ClearParam()
	}
	return
}

// updateAsset -- Updates Asset record from the passed through map data and asset ID
func updateAsset(u SnowAssetStruct, strAssetID string, espXmlmc *apiLib.XmlmcInstStruct, assetType AssetTypeStruct) bool {

	boolRecordUpdated := false
	//Get site ID
	siteID := ""
	siteNameMapping := fmt.Sprintf("%v", APIImportConf.AssetGenericFieldMapping["h_site"])
	siteName := getFieldValue("h_site", siteNameMapping, u, assetType)
	if siteName != "" {
		siteIsInCache, SiteIDCache := siteInCache(siteName)
		//-- Check if we have cached the site already
		if siteIsInCache {
			siteID = strconv.Itoa(SiteIDCache)
		} else {
			siteIsOnInstance, SiteIDInstance := searchSite(siteName, espXmlmc)
			//-- If Returned set output
			if siteIsOnInstance {
				siteID = strconv.Itoa(SiteIDInstance)
			}
		}
	}

	//Get Owned By name
	ownedByName := ""
	ownedByURN := ""
	ownedByMapping := fmt.Sprintf("%v", APIImportConf.AssetGenericFieldMapping["h_owned_by"])
	ownedByID := getFieldValue("h_owned_by", ownedByMapping, u, assetType)
	if ownedByID != "" {
		ownedByIsInCache, ownedByNameCache := customerInCache(ownedByID)
		//-- Check if we have cached the customer already
		if ownedByIsInCache {
			ownedByName = ownedByNameCache
		} else {
			ownedByIsOnInstance, ownedByNameInstance := searchCustomer(ownedByID, espXmlmc)
			//-- If Returned set output
			if ownedByIsOnInstance {
				ownedByName = ownedByNameInstance
			}
		}
	}
	if ownedByName != "" {
		ownedByURN = "urn:sys:0:" + ownedByName + ":" + ownedByID
	}

	lastLoggedOnByURN := ""
	lastLoggedOnUserMapping := fmt.Sprintf("%v", APIImportConf.AssetTypeFieldMapping["h_last_logged_on_user"])
	if lastLoggedOnUserMapping != "" {
		lastLoggedOnByID := getFieldValue("h_last_logged_on_user", lastLoggedOnUserMapping, u, assetType)
		if lastLoggedOnByID != "" {
			lastLoggedOnByIsInCache, lastLoggedOnByNameCache := customerInCache(lastLoggedOnByID)
			//-- Check if we have cached the customer already
			if lastLoggedOnByIsInCache {
				lastLoggedOnByURN = "urn:sys:0:" + lastLoggedOnByNameCache + ":" + lastLoggedOnByID
			} else {
				lastLoggedOnByIsOnInstance, lastLoggedOnByNameInstance := searchCustomer(lastLoggedOnByID, espXmlmc)
				//-- If Returned set output
				if lastLoggedOnByIsOnInstance {
					lastLoggedOnByURN = "urn:sys:0:" + lastLoggedOnByNameInstance + ":" + lastLoggedOnByID
				}
			}
		}
	}

	//Get Used By name
	usedByName := ""
	usedByURN := ""
	usedByMapping := fmt.Sprintf("%v", APIImportConf.AssetGenericFieldMapping["h_used_by"])
	usedByID := getFieldValue("h_owned_by", usedByMapping, u, assetType)
	if usedByID != "" {
		usedByIsInCache, usedByNameCache := customerInCache(usedByID)
		//-- Check if we have cached the customer already
		if usedByIsInCache {
			usedByName = usedByNameCache
		} else {
			usedByIsOnInstance, usedByNameInstance := searchCustomer(usedByID, espXmlmc)
			//-- If Returned set output
			if usedByIsOnInstance {
				usedByName = usedByNameInstance
			}
		}
	}
	if usedByName != "" {
		usedByURN = "urn:sys:0:" + usedByName + ":" + usedByID
	}

	//Get/Set params from map stored against FieldMapping
	strAttribute := ""
	strMapping := ""
	espXmlmc.SetParam("application", appServiceManager)
	espXmlmc.SetParam("entity", "Asset")
	espXmlmc.SetParam("returnModifiedData", "true")
	espXmlmc.OpenElement("primaryEntityData")
	espXmlmc.OpenElement("record")
	espXmlmc.SetParam("h_pk_asset_id", strAssetID)
	espXmlmc.SetParam("h_asset_urn", "urn:sys:entity:com.hornbill.servicemanager:Asset:"+strAssetID)
	//Get asset field mapping
	for k, v := range APIImportConf.AssetGenericFieldMapping {
		strAttribute = fmt.Sprintf("%v", k)
		strMapping = fmt.Sprintf("%v", v)
		if strAttribute == "h_used_by" && usedByName != "" && usedByURN != "" {
			espXmlmc.SetParam("h_used_by", usedByURN)
			espXmlmc.SetParam("h_used_by_name", usedByName)
		}
		if strAttribute == "h_owned_by" && ownedByName != "" && ownedByURN != "" {
			espXmlmc.SetParam("h_owned_by", ownedByURN)
			espXmlmc.SetParam("h_owned_by_name", ownedByName)
		}
		if strAttribute == "h_site" && siteID != "" && siteName != "" {
			espXmlmc.SetParam("h_site", siteName)
			espXmlmc.SetParam("h_site_id", siteID)
		}
		if strAttribute != "h_site" &&
			strAttribute != "h_used_by" &&
			strAttribute != "h_owned_by" &&
			strMapping != "" && getFieldValue(strAttribute, strMapping, u, assetType) != "" {
			espXmlmc.SetParam(strAttribute, getFieldValue(strAttribute, strMapping, u, assetType))
		}
	}

	espXmlmc.CloseElement("record")
	espXmlmc.CloseElement("primaryEntityData")

	var XMLSTRING = espXmlmc.GetParam()

	//-- Check for Dry Run
	if configDryRun != true {
		if configDebug {
			logger(1, "Asset Update XML "+fmt.Sprintf("%s", XMLSTRING), false)
		}
		XMLUpdate, xmlmcErr := espXmlmc.Invoke("data", "entityUpdateRecord")
		if xmlmcErr != nil {
			logger(4, "API Call failed when Updating Asset:"+fmt.Sprintf("%v", xmlmcErr), false)
			logger(1, "API Call XML: "+fmt.Sprintf("%s", XMLSTRING), false)
			mutexCounters.Lock()
			counters.updatedSkipped++
			mutexCounters.Unlock()
			return false
		}

		var xmlRespon xmlmcUpdateResponse

		err := xml.Unmarshal([]byte(XMLUpdate), &xmlRespon)
		if err != nil {
			logger(4, "Unable to read response from Hornbill instance when Updating Asset:"+fmt.Sprintf("%v", err), false)
			logger(1, "API Call XML: "+fmt.Sprintf("%s", XMLSTRING), false)
			mutexCounters.Lock()
			counters.updatedSkipped++
			mutexCounters.Unlock()
			return false
		}
		if xmlRespon.MethodResult != "ok" && xmlRespon.State.ErrorRet != "There are no values to update" {
			logger(4, "Unable to Update Asset: "+xmlRespon.State.ErrorRet, false)
			logger(1, "API Call XML: "+fmt.Sprintf("%s", XMLSTRING), false)
			mutexCounters.Lock()
			counters.updatedSkipped++
			mutexCounters.Unlock()
			return false
		}

		if len(xmlRespon.UpdatedCols.ColList) > 0 {
			boolRecordUpdated = true
		}

		//-- now process extended record data
		espXmlmc.SetParam("application", appServiceManager)
		espXmlmc.SetParam("entity", "Asset")
		espXmlmc.SetParam("returnModifiedData", "true")
		espXmlmc.OpenElement("primaryEntityData")
		espXmlmc.OpenElement("record")
		espXmlmc.SetParam("h_pk_asset_id", strAssetID)
		espXmlmc.CloseElement("record")
		espXmlmc.CloseElement("primaryEntityData")
		//Add extended asset type field mapping
		espXmlmc.OpenElement("relatedEntityData")
		//Set Class & TypeID
		espXmlmc.SetParam("relationshipName", "AssetClass")
		espXmlmc.SetParam("entityAction", "update")
		espXmlmc.OpenElement("record")
		espXmlmc.SetParam("h_pk_asset_id", strAssetID)
		//Get asset field mapping
		for k, v := range APIImportConf.AssetTypeFieldMapping {
			strAttribute = fmt.Sprintf("%v", k)
			strMapping = fmt.Sprintf("%v", v)
			if strAttribute == "h_last_logged_on_user" && lastLoggedOnByURN != "" {
				espXmlmc.SetParam("h_last_logged_on_user", lastLoggedOnByURN)
			}
			if strAttribute != "h_last_logged_on_user" &&
				strMapping != "" && getFieldValue(strAttribute, strMapping, u, assetType) != "" {
				espXmlmc.SetParam(strAttribute, getFieldValue(strAttribute, strMapping, u, assetType))
			}
		}
		espXmlmc.CloseElement("record")
		espXmlmc.CloseElement("relatedEntityData")
		var XMLEXTSTRING = espXmlmc.GetParam()
		if configDebug {
			logger(1, "Asset Extended XML "+fmt.Sprintf("%s", XMLEXTSTRING), false)
		}
		XMLUpdateExt, xmlmcErrExt := espXmlmc.Invoke("data", "entityUpdateRecord")
		if xmlmcErrExt != nil {
			logger(4, "API Call failed when Updating Asset Extended Details:"+fmt.Sprintf("%v", xmlmcErrExt), false)
			logger(1, "API Call XML: "+fmt.Sprintf("%s", XMLEXTSTRING), false)
			mutexCounters.Lock()
			counters.updatedSkipped++
			mutexCounters.Unlock()
			return false
		}
		var xmlResponExt xmlmcUpdateResponse

		err = xml.Unmarshal([]byte(XMLUpdateExt), &xmlResponExt)
		if err != nil {
			logger(4, "Unable to read response from Hornbill instance when Updating Asset Extended Details:"+fmt.Sprintf("%v", err), false)
			logger(1, "API Call XML: "+fmt.Sprintf("%s", XMLSTRING), false)
			mutexCounters.Lock()
			counters.updatedSkipped++
			mutexCounters.Unlock()
			return false
		}
		if xmlResponExt.MethodResult != "ok" && xmlResponExt.State.ErrorRet != "There are no values to update" {
			logger(4, "Unable to Update Asset Extended Details: "+xmlResponExt.State.ErrorRet, false)
			logger(1, "API Call XML: "+fmt.Sprintf("%s", XMLSTRING), false)
			mutexCounters.Lock()
			counters.updatedSkipped++
			mutexCounters.Unlock()
			return false
		}

		if len(xmlResponExt.UpdatedCols.ColList) > 0 {
			boolRecordUpdated = true
		}

		if boolRecordUpdated == false {
			mutexCounters.Lock()
			counters.updatedSkipped++
			mutexCounters.Unlock()
		} else {
			//-- Asset Updated!
			//-- Need to run another update against the Asset for LAST UPDATED and LAST UPDATE BY!
			espXmlmc.SetParam("application", appServiceManager)
			espXmlmc.SetParam("entity", "Asset")
			espXmlmc.OpenElement("primaryEntityData")
			espXmlmc.OpenElement("record")
			espXmlmc.SetParam("h_pk_asset_id", strAssetID)
			espXmlmc.SetParam("h_last_updated", APITimeNow)
			espXmlmc.SetParam("h_last_updated_by", "Import - Update")
			espXmlmc.CloseElement("record")
			espXmlmc.CloseElement("primaryEntityData")
			var XMLSTRING = espXmlmc.GetParam()
			XMLUpdate, xmlmcErr := espXmlmc.Invoke("data", "entityUpdateRecord")
			if xmlmcErr != nil {
				logger(4, "API Call failed when setting Last Updated values:"+fmt.Sprintf("%v", xmlmcErr), false)
				logger(1, "Asset Last Updated XML: "+fmt.Sprintf("%s", XMLSTRING), false)
			}
			var xmlRespon xmlmcResponse
			err := xml.Unmarshal([]byte(XMLUpdate), &xmlRespon)
			if err != nil {
				logger(4, "Unable to read response from Hornbill instance when setting Last Updated values:"+fmt.Sprintf("%v", err), false)
				logger(1, "Asset Last Updated XML: "+fmt.Sprintf("%s", XMLSTRING), false)
			}
			if xmlRespon.MethodResult != "ok" && xmlRespon.State.ErrorRet != "There are no values to update" {
				logger(4, "Unable to set Last Updated details for asset: "+xmlRespon.State.ErrorRet, false)
				logger(1, "Asset Last Updated XML: "+fmt.Sprintf("%s", XMLSTRING), false)
			}
			if configDebug {
				logger(1, "Asset Update XML "+fmt.Sprintf("%s", XMLSTRING), false)
			}
			mutexCounters.Lock()
			counters.updated++
			mutexCounters.Unlock()
		}

	} else {
		//-- Inc Counter
		mutexCounters.Lock()
		counters.updatedSkipped++
		mutexCounters.Unlock()
		logger(1, "Asset Update XML "+fmt.Sprintf("%s", XMLSTRING), false)

		espXmlmc.ClearParam()
	}
	return true
}

// getFieldValue --Retrieve field value from mapping via SQL record map
// k = the Hornbill column ID
// v = the mapping value for the column ID from the config
// u = The Asset record
func getFieldValue(k string, v string, u SnowAssetStruct, assetType AssetTypeStruct) string {
	fieldMap := v
	fieldMap = strings.Replace(fieldMap, `\n`, "\n", -1)
	//-- Match $variable from String
	re1, err := regexp.Compile(`\[(.*?)\]`)
	if err != nil {
		fmt.Printf("[ERROR] %v", err)
	}

	result := re1.FindAllString(fieldMap, 100)
	valFieldMap := ""
	//-- Loop Matches
	for _, val := range result {
		valFieldMap = ""
		valFieldMap = strings.Replace(val, "[", "", 1)
		valFieldMap = strings.Replace(valFieldMap, "]", "", 1)
		boolValFieldMapSet := false
		//Check User Mapping
		for _, userMappings := range APIImportConf.UserMappedFields {
			if userMappings.MappingField == valFieldMap {
				valFieldMap = getSnowUserField(userMappings.SnowField, fmt.Sprintf("%v", u.ID), userMappings, assetType)
				boolValFieldMapSet = true
			}
		}
		if boolValFieldMapSet == false {
			splitString := strings.Split(valFieldMap, ".")
			switch splitString[0] {
			case "Asset":
				valFieldMap, _ = getAssetFieldVal(u, splitString[1])
				break
			case "Hardware":
				valFieldMap, _ = getHardwareFieldVal(u.Hardware, splitString[1])
				break
			case "CustomField":
				valFieldMap = ""
				if len(u.CustomFields) > 0 {
					selectedCustomField, _ := strconv.Atoi(splitString[1])
					valFieldMap, _ = getCustomFieldVal(u.CustomFields[selectedCustomField], splitString[2])
				}
				break
			case "Network":
				valFieldMap = ""
				if len(u.Hardware.NetworkAdapters) > 0 {
					selectedNIC, _ := strconv.Atoi(splitString[1])
					valFieldMap, _ = getNetworkFieldVal(u.Hardware.NetworkAdapters[selectedNIC], splitString[2])
				}
				break
			case "OpticalDrive":
				valFieldMap = ""
				if len(u.Hardware.OpticalDrives) > 0 {
					selectedOpDrive, _ := strconv.Atoi(splitString[1])
					valFieldMap, _ = getOpticalFieldVal(u.Hardware.OpticalDrives[selectedOpDrive], splitString[2])
				} else {

				}
				break
			case "LogicalDisk":
				valFieldMap = ""
				if len(u.Hardware.LogicalDisks) > 0 {
					selectedDisk, _ := strconv.Atoi(splitString[1])
					valFieldMap, _ = getLogicalDiskFieldVal(u.Hardware.LogicalDisks[selectedDisk], splitString[2])
				}
				break
			case "Monitor":
				valFieldMap = ""
				if len(u.Hardware.Monitors) > 0 {
					selectedMonitor, _ := strconv.Atoi(splitString[1])
					valFieldMap, _ = getMonitorFieldVal(u.Hardware.Monitors[selectedMonitor], splitString[2])
				}
				break
			case "DisplayAdapter":
				valFieldMap = ""
				if len(u.Hardware.DisplayAdapters) > 0 {
					selectedDisplayAdapter, _ := strconv.Atoi(splitString[1])
					valFieldMap, _ = getDisplayAdapterFieldVal(u.Hardware.DisplayAdapters[selectedDisplayAdapter], splitString[2])
				}
				break
			}
		}

		fieldMap = strings.Replace(fieldMap, val, valFieldMap, 1)
	}
	return fieldMap
}
