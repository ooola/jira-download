README
======

This program downloads Jira ticket attachments.

```
$ go run ../dlattachments.go -username me@example.com -password MyJiraAPIKey
https://mysite.atlassian.net/secure/attachment/71040/Account+Delete+Email.png --> Account+Delete+Email.832246321.png
https://mysite.atlassian.net/secure/attachment/71278/data_deletion.pdf --> data_deletion.199182812.pdf
https://mysite.atlassian.net/secure/attachment/71038/Chartio+Customer+Deletion.png --> Chartio+Customer+Deletion.287135627.png
...
```

### Notes

To verify a key do

```
curl -v https://mysite.atlassian.net --user 'me@example.com:MyAPIKey'
```

To manually check an issue do:

This works
```
curl -D- \
   -X GET \
   --user 'me@example.com:MyAPIKey' \
   -H "Content-Type: application/json" \
   -o issue \
   "https://mycompany.atlassian.net/rest/api/2/issue/GRC-3461"

curl -D- \
   -X GET \
   --user 'me@example.com:MyAPIKey' \
   -H "Content-Type: application/json" \
   -o foo \
   "https://mycompany.atlassian.net/rest/api/2/search?jql=labels%20IN%20(%22soc2_IRL_fy19%22)"
```

cat issue | jq '.fields.attachment[].content'

cat pretty | jq '.issues[].key'

### Image conversion

Convert PDFs to pngs

find . -name '*.pdf' -exec sh -c 'convert "$0" "${0%.*}.png"' {} \;

Make the video

ffmpeg -pattern_type glob -i "*.png"  video.avi
