<?php

function requestRedirect($curl)
{
    curl_setopt($curl, CURLOPT_TIMEOUT, 10);
    $response = curl_exec($curl);
    $transientErrors = [CURLE_COULDNT_RESOLVE_HOST, CURLE_COULDNT_CONNECT, CURLE_OPERATION_TIMEOUTED];
    if ($response === false && curl_getinfo($curl, CURLINFO_HTTP_CODE) === 0
        && in_array(curl_errno($curl), $transientErrors, true)) {
        $response = curl_exec($curl);
    }
    if ($response === false) {
        throw new RuntimeException('Redirect request failed: ' . curl_error($curl));
    }
    return $response;
}
