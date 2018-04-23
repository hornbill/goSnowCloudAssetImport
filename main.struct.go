package main

import (
	"encoding/xml"
	"sync"
	"time"
)

//----- Constants -----
const version = "1.0.3"
const appServiceManager = "com.hornbill.servicemanager"
const recordsPerPage = 100

//----- Variables -----
var (
	maxLogFileSize      int64
	APIImportConf       apiImportConfStruct
	XmlmcInstanceConfig xmlmcConfig
	Sites               []siteListStruct
	counters            counterTypeStruct
	configFileName      string
	configMaxRoutines   string
	configZone          string
	configDryRun        bool
	configDebug         bool
	Customers           []customerListStruct
	TimeNow             string
	APITimeNow          string
	startTime           time.Time
	endTime             time.Duration
	AssetClass          string
	AssetTypeID         int
	StrAssetType        string
	//espXmlmc            *apiLib.XmlmcInstStruct
	mutex          = &sync.Mutex{}
	mutexBar       = &sync.Mutex{}
	mutexCounters  = &sync.Mutex{}
	mutexCustomers = &sync.Mutex{}
	mutexSite      = &sync.Mutex{}
	worker         sync.WaitGroup
	maxGoroutines  = 1
	logFilePart    = 0
)

//----- Structures -----
type siteListStruct struct {
	SiteName string
	SiteID   int
}
type xmlmcConfig struct {
	instance string
	zone     string
	url      string
}
type counterTypeStruct struct {
	updated        uint16
	created        uint16
	updatedSkipped uint16
	createskipped  uint16
}
type apiImportConfStruct struct {
	AddAssets bool `json:"AddAssets"`
	APIConf   struct {
		APIURL            string `json:"APIURL"`
		ComputersAbstract string `json:"ComputersAbstract"`
		CustomerID        int    `json:"CustomerID"`
		Password          string `json:"Password"`
		UserName          string `json:"UserName"`
		UsersAbstract     string `json:"UsersAbstract"`
	} `json:"APIConf"`
	APIKey                   string `json:"APIKey"`
	AssetGenericFieldMapping map[string]interface{}
	AssetTypeFieldMapping    map[string]interface{}
	AssetIdentifier          string            `json:"AssetIdentifier"`
	AssetTypes               []AssetTypeStruct `json:"AssetTypes"`
	InstanceID               string            `json:"InstanceId"`
	LogSizeBytes             int64             `json:"LogSizeBytes"`
	UpdateAssets             bool              `json:"UpdateAssets"`
	URL                      string
	UserMappedFields         []UserMappedFieldsStruct `json:"UserMappedFields"`
}

//UserMappedFieldsStruct - struct for holding user mapped fields config
type UserMappedFieldsStruct struct {
	MappingField     string `json:"MappingField"`
	SnowField        string `json:"SnowField"`
	Filter           string `json:"Filter"`
	MapField         string `json:"MapField"`
	Order            string `json:"Order"`
	StripDomainRegex string `json:"StripDomainRegex"`
}

//AssetTypeStruct - Struct for asset type data
type AssetTypeStruct struct {
	AssetType       string `json:"AssetType"`
	Description     string `json:"Description"`
	Filter          string `json:"Filter"`
	RecordsAbstract string `json:"RecordsAbstract"`
	UsersAbstract   string `json:"UsersAbstract"`
	AssetID         string `json:"AssetID"`
}
type siteLookupStruct struct {
	Enabled  bool
	QueryCol string
}
type typeLookupStruct struct {
	Enabled   bool
	Attribute string
}
type xmlmcResponse struct {
	MethodResult string       `xml:"status,attr"`
	Params       paramsStruct `xml:"params"`
	State        stateStruct  `xml:"state"`
}

type xmlmcUpdateResponse struct {
	MethodResult string      `xml:"status,attr"`
	UpdatedCols  updatedCols `xml:"params>primaryEntityData>record"`
	State        stateStruct `xml:"state"`
}
type updatedCols struct {
	AssetPK string       `xml:"h_pk_asset_id"`
	ColList []updatedCol `xml:",any"`
}

type updatedCol struct {
	XMLName xml.Name `xml:""`
	Amount  string   `xml:",chardata"`
}

//Site Structs
type xmlmcSiteListResponse struct {
	MethodResult string               `xml:"status,attr"`
	Params       paramsSiteListStruct `xml:"params"`
	State        stateStruct          `xml:"state"`
}
type paramsSiteListStruct struct {
	RowData paramsSiteRowDataListStruct `xml:"rowData"`
}
type paramsSiteRowDataListStruct struct {
	Row siteObjectStruct `xml:"row"`
}
type siteObjectStruct struct {
	SiteID      int    `xml:"h_id"`
	SiteName    string `xml:"h_site_name"`
	SiteCountry string `xml:"h_country"`
}

//----- Customer Structs
type customerListStruct struct {
	CustomerID   string
	CustomerName string
}
type xmlmcCustomerListResponse struct {
	MethodResult      string      `xml:"status,attr"`
	CustomerFirstName string      `xml:"params>firstName"`
	CustomerLastName  string      `xml:"params>lastName"`
	State             stateStruct `xml:"state"`
}

//Asset Structs
type xmlmcAssetResponse struct {
	MethodResult string            `xml:"status,attr"`
	Params       paramsAssetStruct `xml:"params"`
	State        stateStruct       `xml:"state"`
}
type paramsAssetStruct struct {
	RowData paramsAssetRowDataStruct `xml:"rowData"`
}
type paramsAssetRowDataStruct struct {
	Row assetObjectStruct `xml:"row"`
}
type assetObjectStruct struct {
	AssetID    string `xml:"h_pk_asset_id"`
	AssetClass string `xml:"h_class"`
	AssetType  string `xml:"h_country"`
}

//Asset Type Structures
type xmlmcTypeListResponse struct {
	MethodResult string               `xml:"status,attr"`
	Params       paramsTypeListStruct `xml:"params"`
	State        stateStruct          `xml:"state"`
}
type paramsTypeListStruct struct {
	RowData paramsTypeRowDataListStruct `xml:"rowData"`
}
type paramsTypeRowDataListStruct struct {
	Row assetTypeObjectStruct `xml:"row"`
}
type assetTypeObjectStruct struct {
	Type      string `xml:"h_name"`
	TypeClass string `xml:"h_class"`
	TypeID    int    `xml:"h_pk_type_id"`
}
type stateStruct struct {
	Code     string `xml:"code"`
	ErrorRet string `xml:"error"`
}
type paramsStruct struct {
	SessionID string `xml:"sessionId"`
}

//API JSON Structs

//SnowAssetsStruct - struct to store the asset objects in
type SnowAssetsStruct struct {
	Body  []SnowAssetBodyStruct `json:"Body"`
	Links []struct {
		Href   interface{} `json:"Href"`
		Method interface{} `json:"Method"`
		Rel    interface{} `json:"Rel"`
		Title  interface{} `json:"Title"`
	} `json:"Links"`
	Meta []struct {
		Name  string      `json:"Name"`
		Value interface{} `json:"Value"`
	} `json:"Meta"`
}

//SnowAssetBodyStruct - struct to define how we hold the asset body information
type SnowAssetBodyStruct struct {
	Asset SnowAssetStruct `json:"Body"`
	Links []struct {
		Href   interface{} `json:"Href"`
		Method interface{} `json:"Method"`
		Rel    interface{} `json:"Rel"`
		Title  interface{} `json:"Title"`
	} `json:"Links"`
	Meta []interface{} `json:"Meta"`
}

//SnowAssetDetailsStruct - holds more detailed asset information
type SnowAssetDetailsStruct struct {
	Asset SnowAssetStruct `json:"Body"`
}

//SnowAssetStruct - holds asset data
type SnowAssetStruct struct {
	BiosSerialNumber           interface{}             `json:"BiosSerialNumber"`
	ClientConfigurationName    interface{}             `json:"ClientConfigurationName"`
	ClientInstallDate          interface{}             `json:"ClientInstallDate"`
	ClientSiteName             interface{}             `json:"ClientSiteName"`
	ClientVersion              interface{}             `json:"ClientVersion"`
	CustomFields               []SnowCustomFieldStruct `json:"CustomFields"`
	CoreCount                  interface{}             `json:"CoreCount"`
	Domain                     interface{}             `json:"Domain"`
	Hardware                   SnowAssetHardwareStruct `json:"Hardware"`
	HypervisorName             interface{}             `json:"HypervisorName"`
	ID                         interface{}             `json:"Id"`
	IPAddresses                interface{}             `json:"IpAddresses"`
	IsPortable                 interface{}             `json:"IsPortable"`
	IsServer                   interface{}             `json:"IsServer"`
	IsVirtual                  interface{}             `json:"IsVirtual"`
	LastScanDate               interface{}             `json:"LastScanDate"`
	Manufacturer               interface{}             `json:"Manufacturer"`
	MobileDeviceType           interface{}             `json:"MobileDeviceType"`
	Model                      interface{}             `json:"Model"`
	MostFrequentUserID         interface{}             `json:"MostFrequentUserId"`
	MostRecentUserID           interface{}             `json:"MostRecentUserId"`
	Name                       interface{}             `json:"Name"`
	Notes                      interface{}             `json:"Notes"`
	OperatingSystem            interface{}             `json:"OperatingSystem"`
	OperatingSystemServicePack interface{}             `json:"OperatingSystemServicePack"`
	OrgChecksum                interface{}             `json:"OrgChecksum"`
	Organization               interface{}             `json:"Organization"`
	PhoneNumber                interface{}             `json:"PhoneNumber"`
	PhysicalMemory             interface{}             `json:"PhysicalMemory"`
	ProcessorCount             interface{}             `json:"ProcessorCount"`
	ProcessorType              interface{}             `json:"ProcessorType"`
	PurchaseDate               interface{}             `json:"PurchaseDate"`
	PurchaseVendor             interface{}             `json:"PurchaseVendor"`
	PurchasePrice              interface{}             `json:"PurchasePrice"`
	PurchaseCurrency           interface{}             `json:"PurchaseCurrency"`
	PurchaseInvoiceReference   interface{}             `json:"PurchaseInvoiceReference"`
	ScanIdentifier             interface{}             `json:"ScanIdentifier"`
	ScannerVersion             interface{}             `json:"ScannerVersion"`
	Status                     interface{}             `json:"Status"`
	TotalDiskSpace             interface{}             `json:"TotalDiskSpace"`
	TransferDate               interface{}             `json:"TransferDate"`
	UpdatedBy                  interface{}             `json:"UpdatedBy"`
	UpdatedDate                interface{}             `json:"UpdatedDate"`
}

//SnowCustomFieldStruct - holds custom field data
type SnowCustomFieldStruct struct {
	Name  interface{} `json:"Name"`
	Value interface{} `json:"Value"`
}

//SnowAssetHardwareStruct - holds asset hardware data
type SnowAssetHardwareStruct struct {
	BiosDate                   interface{}                      `json:"BiosDate"`
	BiosSerialNumber           interface{}                      `json:"BiosSerialNumber"`
	BiosVersion                interface{}                      `json:"BiosVersion"`
	CoresPerProcessor          interface{}                      `json:"CoresPerProcessor"`
	DisplayAdapters            []SnowAssetDisplayAdaptersStruct `json:"DisplayAdapters"`
	LogicalDisks               []SnowAssetLogicalDisksStruct    `json:"LogicalDisks"`
	MemorySlots                interface{}                      `json:"MemorySlots"`
	MemorySlotsAvailable       interface{}                      `json:"MemorySlotsAvailable"`
	Monitors                   []SnowAssetMonitorsStruct        `json:"Monitors"`
	NetworkAdapters            []SnowAssetNetworkStruct         `json:"NetworkAdapters"`
	NumberOfProcessors         interface{}                      `json:"NumberOfProcessors"`
	OpticalDrives              []SnowAssetOpticalDriveStruct    `json:"OpticalDrives"`
	PhysicalMemoryMb           interface{}                      `json:"PhysicalMemoryMb"`
	ProcessorType              interface{}                      `json:"ProcessorType"`
	SystemDiskSpaceAvailableMb interface{}                      `json:"SystemDiskSpaceAvailableMb"`
	SystemDiskSpaceMb          interface{}                      `json:"SystemDiskSpaceMb"`
	TotalDiskSpaceAvailableMb  interface{}                      `json:"TotalDiskSpaceAvailableMb"`
	TotalDiskSpaceMb           interface{}                      `json:"TotalDiskSpaceMb"`
}

//SnowAssetLogicalDisksStruct - holds logical disk data
type SnowAssetLogicalDisksStruct struct {
	Name       interface{} `json:"Name"`
	SizeMb     interface{} `json:"SizeMb"`
	VolumeName interface{} `json:"VolumeName"`
}

//SnowAssetDisplayAdaptersStruct - holds display adapter data
type SnowAssetDisplayAdaptersStruct struct {
	CurrentResolution interface{} `json:"CurrentResolution"`
	MaxResolution     interface{} `json:"MaxResolution"`
	Name              interface{} `json:"Name"`
}

//SnowAssetNetworkStruct - holds optical drive data
type SnowAssetNetworkStruct struct {
	IPAddress  interface{} `json:"IpAddress"`
	MACAddress interface{} `json:"MacAddress"`
	Name       interface{} `json:"Name"`
}

//SnowAssetOpticalDriveStruct - holds optical drive data
type SnowAssetOpticalDriveStruct struct {
	Name interface{} `json:"Name"`
	Type interface{} `json:"Type"`
}

//SnowAssetMonitorsStruct - holds monitor data
type SnowAssetMonitorsStruct struct {
	MaxResolution interface{} `json:"MaxResolution"`
	Name          interface{} `json:"Name"`
}

//SnowUsersStruct - holds user data
type SnowUsersStruct struct {
	Body []struct {
		User SnowUserDetailsStruct `json:"Body"`
	} `json:"Body"`
}

//SnowUserDetailsStruct - holds a particular users data
type SnowUserDetailsStruct struct {
	CustomerID       interface{} `json:"CustomerId"`
	FullName         interface{} `json:"FullName"`
	ID               interface{} `json:"Id"`
	LastLogon        interface{} `json:"LastLogon"`
	LogonCount       interface{} `json:"LogonCount"`
	UpdatedBy        interface{} `json:"UpdatedBy"`
	UpdatedDate      interface{} `json:"UpdatedDate"`
	UserIsAccessible interface{} `json:"UserIsAccessible"`
	Username         interface{} `json:"Username"`
}
