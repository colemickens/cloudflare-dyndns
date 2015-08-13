param(
    [string] $oldIp = "",
    [string] $newIp = "",
    [string] $email = "",
    [string] $key = ""
)

$authHeaders = @{
    "X-Auth-Email" = $email;
    "X-Auth-Key" = $key }

$zoneResponseRaw = Invoke-WebRequest -Method Get -Uri "https://api.cloudflare.com/client/v4/zones" -Headers  $authHeaders
$zoneResponse = ConvertFrom-Json ($zoneResponseRaw).Content

$zoneResponse.result | % {
    $zoneId = $_.id

    $recordResponse = ConvertFrom-Json (Invoke-WebRequest `
        -Uri "https://api.cloudflare.com/client/v4/zones/$zoneId/dns_records" `
        -Method Get -Headers  $authHeaders)
    
    $recordResponse.result | % {
        $recordId = $_.id
        if ($_.content -ne $oldIp)
        {
            New-Object psobject -Property @{ "name" = $_.name; "response" = $_.content; "action" = "skipped" }
            return
        }
        elseif ($_.content -eq $oldIp)
        {
            $updateHeaders = $authHeaders.Clone()
            $updateHeaders += @{"Content-Type" = "application/json"}
            try {
            $updateResponseRaw = Invoke-WebRequest `
                -Uri "https://api.cloudflare.com/client/v4/zones/$zoneId/dns_records/$recordId" `
                -Method Put -Headers  $updateHeaders `
                -Body (ConvertTo-Json `
                    @{ `
                        "id" = "$recordId"; `
                        "type" = $_.type; `
                        "name" = $_.name;
                        "content" = $newIp; `
                    })
            } catch {
                $exceptionStream = $_.Exception.Response.GetResponseStream()
                $exceptionText = (New-Object System.IO.StreamReader($exceptionStream)).ReadToEnd();
                throw $exceptionText
            }

            $updateResponse = (ConvertFrom-Json $updateResponseRaw).result

            New-Object psobject -Property @{ "name" = $_.name; "response" = $updateResponse.content; "action" = "updated" }
        }
    }
} | Format-Table name,response,action
