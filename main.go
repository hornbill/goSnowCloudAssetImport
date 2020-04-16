package main

//----- Packages -----
import (
	"flag"
	"fmt"
	"strconv"
	"strings"

	"github.com/hornbill/color"

	"time"
)

//----- Main Function -----
func main() {
	//-- Start Time for Durration
	startTime = time.Now()
	//-- Start Time for Log File
	TimeNow = time.Now().Format(time.RFC3339)
	APITimeNow = strings.Replace(TimeNow, "T", " ", 1)
	APITimeNow = strings.Replace(APITimeNow, "Z", "", 1)
	//-- Remove :
	TimeNow = strings.Replace(TimeNow, ":", "-", -1)
	//-- Grab Flags
	flag.StringVar(&configFileName, "file", "conf.json", "Name of Configuration File To Load")
	flag.BoolVar(&configDryRun, "dryrun", false, "Allow the Import to run without Creating or Updating Assets")
	flag.BoolVar(&configDebug, "debug", false, "Produces extended logging.")
	flag.StringVar(&configMaxRoutines, "concurrent", "1", "Maximum number of Assets to import concurrently.")
	flag.BoolVar(&configVersion, "version", false, "Return the Version number")
	//-- Parse Flags
	flag.Parse()

	//-- If configVersion just output version number and die
	if configVersion {
		fmt.Printf("%v \n", version)
		return
	}

	//-- Output
	logger(1, "---- Hornbill Snow License Manager Cloud Asset Import Utility V"+version+" ----", true)
	logger(1, "Flag - Config File "+fmt.Sprintf("%s", configFileName), true)
	logger(1, "Flag - Dry Run "+fmt.Sprintf("%v", configDryRun), true)

	//Check maxGoroutines for valid value
	maxRoutines, err := strconv.Atoi(configMaxRoutines)
	if err != nil {
		color.Red("Unable to convert maximum concurrency of [" + configMaxRoutines + "] to type INT for processing")
		return
	}
	maxGoroutines = maxRoutines

	if maxGoroutines < 1 || maxGoroutines > 3 {
		color.Red("The maximum concurrent assets allowed is between 1 and 3 (inclusive).\n\n")
		color.Red("You have selected " + configMaxRoutines + ". Please try again, with a valid value against ")
		color.Red("the -concurrent switch.")
		return
	}

	//--
	//-- Load Configuration File Into Struct
	APIImportConf = loadConfig()
	if APIImportConf.LogSizeBytes > 0 {
		maxLogFileSize = APIImportConf.LogSizeBytes
	}

	for _, v := range APIImportConf.AssetTypes {
		skipRows := 0
		totalAssets := 0
		boolEndLoop := false
		//pageNum := 0
		StrAssetType := fmt.Sprintf("%v", v.AssetType)
		//Set Asset Class & Type vars from instance
		AssetClass, AssetTypeID = getAssetClass(StrAssetType)

		for boolEndLoop == false {
			objAssets := getAssets(v, skipRows)
			skipRows += recordsPerPage
			if totalAssets == 0 {
				//Get Total  Assets Count
				for _, v := range objAssets.Meta {
					if v.Name == "Count" {
						totalAssets, _ = strconv.Atoi(fmt.Sprintf("%v", v.Value))
					}
				}
				logger(1, "Total ["+v.Description+"] assets found: "+strconv.Itoa(totalAssets), true)
			}

			if len(objAssets.Body) <= 0 {
				boolEndLoop = true
			} else {
				processAssets(objAssets.Body, v)
			}
		}
	}

	//-- End output
	if APIImportConf.UpdateAssets {
		logger(1, "Updated: "+fmt.Sprintf("%d", counters.updated), true)
		logger(1, "Updated Skipped: "+fmt.Sprintf("%d", counters.updatedSkipped), true)
	}
	if APIImportConf.AddAssets {
		logger(1, "Created: "+fmt.Sprintf("%d", counters.created), true)
		logger(1, "Created Skipped: "+fmt.Sprintf("%d", counters.createskipped), true)
	}
	//-- Show Time Takens
	endTime = time.Now().Sub(startTime)
	logger(1, "Time Taken: "+fmt.Sprintf("%v", endTime), true)
	logger(1, "---- Hornbill Snow License Manager Cloud Asset Import Complete ---- ", true)
}
