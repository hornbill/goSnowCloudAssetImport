### Snow License Manager (Cloud) Asset Import Go - [GO](https://golang.org/) Import Script to Hornbill

### Quick links
- [Installation](#installation)
- [Configuration](#config)
    - [Instance Config](#instanceconfig)
    - [SQL Config](#sqlconf)
    - [Asset Types](#assettypes)
    - [Asset Field Mapping (Generic)](#assetgenericfieldmapping)
    - [Asset Field Mapping (Type Specific)](#assettypefieldmapping)
    - [Mapping Fields from associated Users](#usermappedfields)
- [Asset Data Mapping](#mapping)
- [Execute](#execute)
- [Testing](testing)
- [Scheduling](#scheduling)
- [Logging](#logging)
- [Error Codes](#errors)

# Installation

#### Windows
* Download the archive containing the import executable
* Extract zip into a folder you would like the application to run from e.g. `C:\asset_import\`
* Open '''conf.json''' and add in the necessary configration
* Open Command Line Prompt as Administrator
* Change Directory to the folder with goSnowCloudAssetImport_x64.exe `C:\asset_import\`
* Run the command :
* For Windows 32bit Systems: goSnowCloudAssetImport_x86.exe -dryrun=true -file=conf.json
* For Windows 64bit Systems: goSnowCloudAssetImport_x64.exe -dryrun=true -file=conf.json

# Config

Example JSON File:

```json
{
    "APIKey": "Hornbill API Key goes here",
    "InstanceId": "Hornbill Instance ID goes here",
    "AssetIdentifier": "h_name",
    "LogSizeBytes": 1000000,
    "AddAssets": true,
    "UpdateAssets": true,
    "APIConf": {
        "APIURL": "Snow License Manager API URL goes here",
        "UserName": "username",
        "Password": "password",
        "CustomerID": 123
    },
    "AssetTypes": [
        {
            "AssetType": "Server",
            "Filter":"OperatingSystem eq 'VMware ESXi'",
            "Description":"All ESX Hosts",
            "RecordsAbstract": "computers",
            "UsersAbstract": "users",
            "AssetID": "Asset.Name"
        },
        {
            "AssetType": "Server",
            "Filter":"indexof(OperatingSystem, 'Windows Server') ge 0 and IsVirtual eq false",
            "Description":"Hardware Servers - non ESX",
            "RecordsAbstract": "computers",
            "UsersAbstract": "users",
            "AssetID": "Asset.Name"
        },
        {
            "AssetType":"Virtual Machine",
            "Filter":"indexof(OperatingSystem, 'Windows Server') ge 0 and IsVirtual eq true",
            "Description":"Virtual Windows Servers",
            "RecordsAbstract": "computers",
            "UsersAbstract": "users",
            "AssetID": "Asset.Name"
        },
        {
            "AssetType":"Desktop",
            "Filter":"Status eq 'Active' and Manufacturer eq 'Dell Inc.' and (indexof(Model, 'OptiPlex') ge -1 or indexof(Model, 'Inspiron') ge -1)",
            "Description":"Active Dell PCs",
            "RecordsAbstract": "computers",
            "UsersAbstract": "users",
            "AssetID": "Asset.Name"
        },
        {
            "AssetType":"Laptop",
            "Filter":"Status eq 'Active' and Manufacturer eq 'Dell Inc.' and indexof(Model, 'Latitude') ge 0",
            "Description":"Active Laptops",
            "RecordsAbstract": "computers",
            "UsersAbstract": "users",
            "AssetID": "Asset.Name"
        }
    ],
    "AssetGenericFieldMapping": {
        "h_name": "[Asset.Name]",
        "h_site": "[Asset.ClientSiteName]",
        "h_asset_tag": "[Asset.Name]",
        "h_acq_method": "",
        "h_actual_retired_date": "",
        "h_beneficiary": "",
        "h_building": "",
        "h_cost": "",
        "h_cost_center": "",
        "h_country": "",
        "h_created_date": "",
        "h_deprec_method": "",
        "h_deprec_start": "",
        "h_description": "Snow Asset ID: [Asset.ID]",
        "h_disposal_price": "",
        "h_disposal_reason": "",
        "h_floor": "",
        "h_geo_location": "",
        "h_invoice_number": "",
        "h_location": "",
        "h_location_type": "",
        "h_maintenance_cost": "",
        "h_maintenance_ref": "",
        "h_notes": "[Asset.Notes]\\nDisplay Adapter: [DisplayAdapter.0.Name]\\nLogical Disks One: [LogicalDisk.0.Name] [LogicalDisk.0.SizeMb] [LogicalDisk.0.VolumeName]",
        "h_operational_state": "",
        "h_order_date": "",
        "h_order_number": "[Asset.PurchaseInvoiceReference]",
        "h_owned_by": "[MostFrequentUserId]",
        "h_product_id": "",
        "h_received_date": "",
        "h_residual_value": "",
        "h_room": "",
        "h_scheduled_retire_date": "",
        "h_supplier_id": "",
        "h_supported_by": "",
        "h_used_by": "[MostFrequentUserId]",
        "h_version": "",
        "h_warranty_expires": "",
        "h_warranty_start": ""
    },
    "AssetTypeFieldMapping": {
        "h_name": "[Asset.Name]",
        "h_mac_address": "[Network.0.MACAddress]",
        "h_net_ip_address": "[Network.0.IPAddress]",
        "h_net_computer_name": "[Asset.Name]",
        "h_net_win_domain": "[Asset.Domain]",
        "h_model": "[Asset.Model]",
        "h_manufacturer": "[Asset.Manufacturer]",
        "h_cpu_info": "[Hardware.ProcessorType]",
        "h_description": "",
        "h_last_logged_on": "[MostRecentUserLastLogonDate]",
        "h_last_logged_on_user": "[MostRecentUserId]",
        "h_memory_info": "[Hardware.PhysicalMemoryMb]",
        "h_net_win_dom_role": "",
        "h_optical_drive": "[OpticalDrive.0.Name] [OpticalDrive.0.Type]",
        "h_os_description": "[Asset.OperatingSystem]",
        "h_os_registered_to": "",
        "h_os_serial_number": "",
        "h_os_service_pack": "[Asset.OperatingSystemServicePack]",
        "h_os_type": "",
        "h_os_version": "",
        "h_physical_disk_size": "[Hardware.TotalDiskSpaceMb]",
        "h_serial_number": "",
        "h_cpu_clock_speed": "",
        "h_physical_cpus": "[Asset.ProcessorCount]",
        "h_logical_cpus": "[Asset.CoresPerProcessor]",
        "h_bios_name": "",
        "h_bios_manufacturer": "",
        "h_bios_serial_number": "[Hardware.BiosSerialNumber]",
        "h_bios_release_date": "[Hardware.BiosDate]",
        "h_bios_version": "",
        "h_max_memory_capacity": "",
        "h_number_memory_slots": "[Hardware.MemorySlots]",
        "h_net_name": "",
        "h_subnet_mask": ""
    },
    "UserMappedFields":[
        {
            "MappingField":"MostRecentUserId",
            "SnowField":"Username",
            "StripDomainRegex":".{0,8}\\\\",
            "Filter":"FullName ne null",
            "Order":"LastLogon desc"
        },
        {
            "MappingField":"MostFrequentUserId",
            "SnowField":"Username",
            "StripDomainRegex":".{0,8}\\\\",
            "Filter":"FullName ne null",
            "Order":"LogonCount desc"
        },
        {
            "MappingField":"MostRecentUserLastLogonDate",
            "SnowField":"LastLogon",
            "StripDomainRegex":"",
            "Filter":"FullName ne null",
            "Order":"LogonCount desc"
        }
    ]
}
```

#### InstanceConfig
* **APIKey** - A Hornbill API key for a user account with the correct permissions to carry out all of the required API calls
* **InstanceId** - Case-sensitive ID of the Hornbill instance to import to
* **AssetIdentifier** - The column in the asset table in Hornbill that holds the unique asset identifier (so that the code can work out which asset records are to be inserted or updated)
* **LogSizeBytes** - The maximum size that the generated Log Files should be, in bytes. Setting this value to 0 will cause the tool to create one log file only and not split the results between multiple logs
* **AddAssets** - Boolean true or false, defines wether discovered assets should be added to Hornbill
* **UpdateAssets**  - Boolean true or false, defines wether discovered assets that already exist in Hornbill should be updated

#### APIConf
* **APIURL** - The URL of the Snow License Server Cloud API
* **UserName** - The username for the Snow account to use to connect to the Snow License Server Cloud API
* **Password** - The password for above UserName
* **CustomerID** - An integer value, the Customer ID of your Snow account

#### AssetTypes
An array of objects that define the asset types you are importing, and information about how to identify and import these records in to Hornbill. Each object should contain the following properties:
* **AssetType** - The Hornbill asset type you are importing
* **Filter** - The filter to use in the Snow API call to retrieve asset records
* **Description** - A text description to highlight & identify at a glance which asset type is being imported
* **RecordsAbstract** - The Snow abstract name to use for the particular asset type import (computers, mobiledevices etc)
* **UsersAbstract** - The Snow abstract name to use for retrieving asset user imformation (users)
* **AssetID** - The column name and storage object from Snow that holds a matching unique identifier to the Hornbill column defined above (InstanceConfig > AssetIdentifier)

#### AssetGenericFieldMapping
This section allows the mapping data retrieved from the Snow API in to the generic Asset record. 
Any value wrapped with [] will be populated with the corresponding response from the Snow API call, and any Other Value is treated literally as written examples below.
Any property value can contain any mixture of mapped vales and plain text. More than one mapped value can be added to a property value, where appropriate.
* "h_name":"[Asset.Name]" - the value of Asset.Name is taken from the API output and populated within this field
* "h_description":"This is a description" - the value of "h_description" would be populated with "This is a description" for ALL imported assets
* "h_notes": "Display Adapter: [DisplayAdapter.0.Name]\\nLogical Disks One: [LogicalDisk.0.Name] [LogicalDisk.0.SizeMb]" - An example of using plain text and multiple mapped values within one property. Note, \\n will add a line break in to your imported string, should the destination field support it
* "h_site":"[SiteName]", - When a string is passed to the h_site field, the script attempts to resolve the given site name against the Site entity, and populates this (and h_site_id) with the correct site information. If the site cannot be resolved, the site details are not populated for the Asset record being imported.
* "h_owned_by":"[MostFrequentUserId]" - when a valid Hornbill User ID (for a Full or Basic User) can be resolved by this mapping, the user is verified on your Hornbill instance, and the tool will complete the h_owned_by and h_owned_by_name columns appropriately. See **UserMappedFields** section for more information.
* "h_used_by":"[MostFrequentUserId]" - when a valid Hornbill User ID (for a Full or Basic User) can be resolved by this mapping, the user is verified on your Hornbill instance, and the tool will complete the h_used_by and h_used_by_name columns appropriately. See **UserMappedFields** section for more information.
* "h_last_logged_on":"[MostRecentUserLastLogonDate]" - when a valid Hornbill User ID (for a Full or Basic User) can be resolved by this mapping, the user is verified on your Hornbill instance, and the tool will complete the h_last_logged_on column with an appropriate URN. See **UserMappedFields** section for more information.

#### AssetTypeFieldMapping
* Maps data in to the type-specific Asset record, using the same rules as AssetGenericFieldMapping

#### UserMappedFields
An array of objects, where each will define the column to retrieve from the API call, and the rules to retrieve the record, from the list of users associated to an asset. A limit of 1 will be automatically applied to the resultset, so only one user record will ever be returned.
* **MappingField** The name of the mapping field, that when populated can be used in either AssetGenericFieldMapping or AssetTypeFieldMapping column mappings
* **SnowField** - The field to return from the Snow API call 
* **StripDomainRegex** - A regular expression that, when matched, will be stripped from the field being returned. Particularly useful to remove the domain information from a returned user name
* ** Filter** - The filter to use in the API call then returning the list of users that have been logged in to the asset
* **Order** - The order in which to return the users.

# Mapping
We've provided the following means of mapping complex data objects and properties from the Snow API calls in to your Hornbill asset columns.
When mapping simple string data from Snow in to Hornbill, the Snow object name and property name should be seperated by a full stop.
When mapping array data from Snow in to Hornbill, the Snow object name, array entry value and property name should all be seperated by a full stop.
Both of the above should be wrapped in [square.braces].
The following objects and properties have been provided for use in your mappings:
- **Asset** - Single value string data, from the primary asset data. Shared object between different asset abstracts. Any of these values can be accessed using, for example, [Asset.Name] or [Asset.Manufacturer].
    - BiosSerialNumber - From the computers abstract
	- ClientConfigurationName - From the computers abstract
	- ClientInstallDate - From the computers abstract
	- ClientSiteName - From the computers abstract
	- ClientVersion - From the computers abstract
	- CoreCount - From the computers abstract
	- Domain - From the computers abstract
	- HypervisorName - From the computers abstract
	- ID - Primary Key - From both the computers and mobiledevices abstracts.
	- IPAddresses - From the computers abstract
	- IsPortable - From the computers abstract
	- IsServer - From the computers abstract
	- IsVirtual - From the computers abstract
	- LastScanDate - From both the computers and mobiledevices abstracts.
	- Manufacturer - From both the computers and mobiledevices abstracts.
	- MobileDeviceType - From the mobiledevices abstract.
	- Model - From both the computers and mobiledevices abstracts.
	- MostFrequentUserID - From the computers abstract
	- MostRecentUserID - From the computers abstract
	- Name - From both the computers and mobiledevices abstracts.
	- Notes - From the computers abstract
	- OperatingSystem - From both the computers and mobiledevices abstracts.
	- OperatingSystemServicePack - From the computers abstract
	- OrgChecksum - From the computers abstract
	- Organization - From both the computers and mobiledevices abstracts.
	- PhoneNumber - From the mobiledevices abstract.
	- PhysicalMemory - From the computers abstract
	- ProcessorCount - From the computers abstract
	- ProcessorType - From the computers abstract
	- PurchaseDate - From the computers abstract
	- PurchaseVendor - From the computers abstract
	- PurchasePrice - From the computers abstract
	- PurchaseCurrency - From the computers abstract
	- PurchaseInvoiceReference - From the computers abstract
	- ScanIdentifier - From the computers abstract
	- ScannerVersion - From the computers abstract
	- Status - From the computers abstract
	- TotalDiskSpace - From the computers abstract
	- TransferDate - From the computers abstract
	- UpdatedBy - From both the computers and mobiledevices abstracts.
	- UpdatedDate - From both the computers and mobiledevices abstracts.
- **Hardware** - Information specific to the asset hardware, specifically for the computers abstract. Any of these values can be accessed using [Hardware.ProcessorType] for example.
    - BiosDate
	- BiosSerialNumber
	- BiosVersion
	- CoresPerProcessor
	- MemorySlots
	- MemorySlotsAvailable
	- NumberOfProcessors
	- PhysicalMemoryMb
	- ProcessorType
	- SystemDiskSpaceAvailableMb
	- SystemDiskSpaceMb
	- TotalDiskSpaceAvailableMb
	- TotalDiskSpaceMb
- **CustomField** - Information from the CustomFields stored within an asset in Snow. This is a complex data type rather than simple, so is stored in an array. Any values from this array can be accessed using, for example, [CustomField.0.Value] to retrieve the Value of the first (0) entry in the CustomField array, or [CustomField.2.Name] to retrieve the name of the third custom field in the array.
    - Name
    - Value
- **Network** - Information from the Network devices attached to an asset in Snow. This is a complex data type rather than simple, so is stored in an array. Any values from this array can be accessed using, for example, [Network.0.IPAddress] to retrieve the IP address of the primary network device in the asset, or [Network.1.Name] to retrieve the name of the secondary network device in the asset.
	- IPAddress
	- MACAddress
	- Name
- **LogicalDisk** - Information from the Logical Disk Drives attached to an asset in Snow. This is a complex data type rather than simple, so is stored in an array. Any values from this array can be accessed using, for example, [LogicalDisk.0.SizeMb] to retrieve the size in MB of the first logical disk in the asset, or [LogicalDisk.1.VolumeName] to retrieve the volume name of the second logical drive in the asset.
	- Name
	- SizeMb
	- VolumeName
- **OpticalDrive** - Information from the Optical Drives attached to an asset in Snow. This is a complex data type rather than simple, so is stored in an array. Any values from this array can be accessed using, for example, [OpticalDrive.0.Name] to retrieve the name of the first optical disk in the asset, or [OpticalDrive.1.Type] to retrieve the type of the second optical drive in the asset.
    - Name
    - Type
- **Monitor** - Information from the Monitors attached to an asset in Snow. This is a complex data type rather than simple, so is stored in an array. Any values from this array can be accessed using, for example, [Monitor.0.MaxResolution] to retrieve the maximum resolution of the first monitor connected to the asset, or [Monitor.1.Name] to retrieve the name of the second monitor connected to the asset.
    - Name
    - MaxResolution
- **DisplayAdapter** - Information from the Display Adapters attached to an asset in Snow. This is a complex data type rather than simple, so is stored in an array. Any values from this array can be accessed using, for example, [DisplayAdapter.0.MaxResolution] to retrieve the maximum resolution of the first display adapter connected to the asset, or [DisplayAdapter.1.MaxResolution] to retrieve the name of the second display adapter connected to the asset.
	- CurrentResolution
	- MaxResolution
	- Name

# Execute
Command Line Parameters
* file - Defaults to `conf.json` - Name of the Configuration file to load
* debug - Defaults to `false` - Set to true to enable additional logging (outputs all API call payload for example)
* dryrun - Defaults to `false` - Set to true and the XMLMC for Create and Update assets will not be called and instead the XML will be dumped to the log file, this is to aid in debugging the initial connection information.
* zone - Defaults to `eur` - Allows you to change the ZONE used for creating the XMLMC EndPoint URL https://{ZONE}api.hornbill.com/{INSTANCE}/
* concurrent - defaults to `1`. This is to specify the number of assets that should be imported concurrently, and can be an integer between 1 and 3 (inclusive). 1 is the slowest level of import, but does not affect performance of your Hornbill instance, and 3 will process the import much more quickly but could affect instance performance while the import is running.

# Testing
If you run the application with the argument dryrun=true then no assets will be created or updated, the XML used to create or update will be saved in the log file so you can make sure that the data mappings are correct before running the import.

'goSnowCloudAssetImport_x64.exe -dryrun=true'

# Scheduling

### Windows
You can schedule goSnowCloudAssetImport_x64.exe to run with any optional command line argument from Windows Task Scheduler.
* Ensure the user account running the task has rights to goSnowCloudAssetImport_x64.exe and the containing folder.
* Make sure the Start In parameter contains the folder where goSnowCloudAssetImport_x64.exe resides in otherwise it will not be able to pick up the correct path.

# Logging
All Logging output is saved in the log directory in the same directory as the executable the file name contains the date and time the import was run 'Asset_Import_2015-11-06T14-26-13Z.log'

# Errors
* `100` - Unable to create log File
* `101` - Unable to create log folder
* `102` - Unable to Load Configuration File
