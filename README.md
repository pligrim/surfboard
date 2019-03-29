# surfboard
Helm chart plotter
Creates a simple requirements map as an HTML file, with the option to include release notes

To Run: go run Surfboard repo/chartName chartVersion

To Include Release Notes: go run Surfboard -notes repo/chartName chartVersion
Release notes show all historic releases, listing the Jira ticket numbers in each release

Version Routing
the option -routes can be used to include version routing information in the output

note the chart must exist in your chart museum

## Posting to Confluence 
python create_Confluence_Page.py Space_ID ParentPage_ID ProjectName

  python create_Confluence_Page.py SC 127644266 Status
  
Requires Environment  Variables:
    * CONF_USER
    * CONF_PASSWORD
