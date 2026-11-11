<?php
function checkCurl($condition, $message) {
    if (!$condition) {
        throw new RuntimeException($message);
    }
}

function checkSingleTransfers($url) {
    $handle = curl_init($url);
    checkCurl(curl_setopt_array($handle, [CURLOPT_POST => true, CURLOPT_RETURNTRANSFER => true]), 'setopt_array failed');
    $expected = curl_exec($handle);
    checkCurl(is_string($expected), 'Return-transfer did not return a string');
    checkCurl(curl_getinfo($handle, CURLINFO_RESPONSE_CODE) === 200, 'Response code changed');
    $copy = curl_copy_handle($handle);
    checkCurl(curl_exec($copy) === $expected, 'Copied handle changed response');
    checkCurl(curl_exec($handle) === $expected, 'Reused handle changed response');

    $chunks = '';
    $headers = '';
    checkCurl(curl_setopt($handle, CURLOPT_WRITEFUNCTION, function ($handle, $bytes) use (&$chunks) {
        $chunks .= $bytes;
        return strlen($bytes);
    }), 'Write callback setup failed');
    checkCurl(curl_setopt($handle, CURLOPT_HEADERFUNCTION, function ($handle, $bytes) use (&$headers) {
        $headers .= $bytes;
        return strlen($bytes);
    }), 'Header callback setup failed');
    checkCurl(curl_exec($handle) === true, 'Callback transfer did not return true');
    checkCurl($chunks === $expected, 'Write callback received missing or duplicate bytes');
    checkCurl(strpos($headers, 'Content-Type: application/json') !== false, 'Header callback lost headers');

    $copy = curl_copy_handle($handle);
    $chunks = '';
    checkCurl(curl_exec($copy) === true && $chunks === $expected, 'Copied callback changed delivery');

    curl_reset($handle);
    $file = tmpfile();
    checkCurl(curl_setopt_array($handle, [CURLOPT_URL => $url, CURLOPT_POST => true, CURLOPT_FILE => $file]), 'File sink setup failed');
    checkCurl(curl_exec($handle) === true, 'File transfer did not return true');
    rewind($file);
    checkCurl(stream_get_contents($file) === $expected, 'File sink changed response');
    fclose($file);

    curl_reset($handle);
    curl_setopt_array($handle, [CURLOPT_URL => $url, CURLOPT_POST => true]);
    ob_start();
    $result = curl_exec($handle);
    $stdout = ob_get_clean();
    checkCurl($result === true && $stdout === $expected, 'Default output changed response');

    curl_setopt($handle, CURLOPT_WRITEFUNCTION, function ($handle, $bytes) {
        return 0;
    });
    checkCurl(curl_exec($handle) === false, 'Short callback return did not abort transfer');
    checkCurl(curl_errno($handle) === CURLE_WRITE_ERROR, 'Abort error code changed');

    $exception = new RuntimeException('Application callback error');
    curl_setopt($handle, CURLOPT_WRITEFUNCTION, function ($handle, $bytes) use ($exception) {
        throw $exception;
    });
    try {
        curl_exec($handle);
        throw new RuntimeException('Callback exception was swallowed');
    } catch (RuntimeException $caught) {
        checkCurl($caught === $exception, 'Callback exception changed');
    }
    return $expected;
}

function checkMultiTransfers($url, $expected) {
    $multi = curl_multi_init();
    $handles = [];
    $callbackBodies = ['', ''];
    foreach (['return', 'callback', 'callback', 'failed'] as $index => $mode) {
        $curl = curl_init($mode === 'failed' ? 'http://127.0.0.1:1/other' : $url);
        curl_setopt_array($curl, [CURLOPT_POST => true, CURLOPT_RETURNTRANSFER => true, CURLOPT_CONNECTTIMEOUT => 1]);
        if ($mode === 'callback') {
            $bodyIndex = $index - 1;
            curl_setopt($curl, CURLOPT_WRITEFUNCTION, function ($handle, $bytes) use (&$callbackBodies, $bodyIndex) {
                $callbackBodies[$bodyIndex] .= $bytes;
                return strlen($bytes);
            });
        }
        checkCurl(curl_multi_add_handle($multi, $curl) === CURLM_OK, 'Multi add failed');
        $handles[] = $curl;
    }
    do {
        $result = curl_multi_exec($multi, $running);
        checkCurl($result === CURLM_OK, 'Multi execution failed');
        if ($running) {
            curl_multi_select($multi, 1);
        }
    } while ($running);

    $completed = [];
    while ($message = curl_multi_info_read($multi, $queued)) {
        checkCurl($message['msg'] === CURLMSG_DONE, 'Completion message changed');
        $index = array_search($message['handle'], $handles, true);
        checkCurl($index !== false && !isset($completed[$index]), 'Unknown or duplicate completion');
        $completed[$index] = $message['result'];
    }
    checkCurl(count($completed) === 4, 'Multi completion queue was consumed');
    foreach ([0, 1, 2] as $index) {
        checkCurl($completed[$index] === CURLE_OK, 'Successful completion changed');
    }
    checkCurl($completed[3] === CURLE_COULDNT_CONNECT, 'Failed completion changed');
    checkCurl(curl_multi_getcontent($handles[0]) === $expected, 'Multi return buffer changed');
    checkCurl($callbackBodies === [$expected, $expected], 'Multi callbacks changed response');
    foreach ($handles as $curl) {
        checkCurl(curl_multi_remove_handle($multi, $curl) === CURLM_OK, 'Multi remove failed');
    }
    return count($completed);
}

$url = $_GET['provider'].'/other';
$expected = checkSingleTransfers($url);
$completions = checkMultiTransfers($url, $expected);

header('Content-Type: application/json');
echo json_encode(['ok' => true, 'bodyHash' => hash('sha256', $expected), 'completions' => $completions]);
