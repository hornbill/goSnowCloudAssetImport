# CHANGELOG

## 1.1.1 (April 16th, 2020)

Changes:

- Updated code to support Core application and platform changes
- Added version flag to enable auto-build

## 1.1.0 (December 11th 2018)

Features:

- Added support for populating the company fields against an asset. The tool will perform a Company look-up if a company name (in the h_company__name mapping) has been provided, before populating the company name and ID fields against the new or updated asset
- Removed need to provide zone CLI parameter

## 1.0.3 (April 23rd 2018)

Feature:

- Added account verification and URN building when value supplied to h_last_logged_on_user column

## 1.0.2 (April 13th 2018)

Feature:

- Added flags to switch on/off the adding or updating of assets

## 1.0.1 (March 22nd 2018)

Defect fix:

- Fixed issue with dryrun mode not outputting to log when assets detected on Hornbill instance

## 1.0.0 (March 21st 2018)

Features:

- Initial Release
