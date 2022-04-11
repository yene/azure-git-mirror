# Azure Devops Git Mirror
Mirros all Git repos from your Azure Devops account. Moves deleted to archive folder.
Linux build can be found in the repo. Not tested on Windows.

Use `-wiki` to also download project wikis.
If you get an error please check your PAT permissions.

## Features & TODOs
- [x] mirror Git accounts (git clone, then git pull)
- [x] handle empty repos
- [x] move deleted repos to `archive` folder
- [x] summary
- [x] Downloading project wikis with `-wiki`
- [ ] handle expired PAT
- [ ] handle PAT with missing permissions

## Setup PAT token
* [Use personal access tokens](https://docs.microsoft.com/en-us/azure/devops/organizations/accounts/use-personal-access-tokens-to-authenticate?view=azure-devops&tabs=preview-page)

## Links
* code samples https://github.com/microsoft/azure-devops-dotnet-samples/tree/master/ClientLibrary/Samples/Git
*
