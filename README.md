# Azure Devops Git Mirror
Mirros all Git repos from your Azure Devops account. Moves deleted to archive folder.
Linux build can be found in the repo. Not tested on Windows.

## Features
- [x] mirror Git accounts (git clone, then git pull)
- [x] handle empty repos
- [x] move deleted repos to `archive` folder
- [x] summary
- [ ] handle expired PAT

## Setup PAT token
* [Use personal access tokens](https://docs.microsoft.com/en-us/azure/devops/organizations/accounts/use-personal-access-tokens-to-authenticate?view=azure-devops&tabs=preview-page)

## Links
* code samples https://github.com/microsoft/azure-devops-dotnet-samples/tree/master/ClientLibrary/Samples/Git
*
