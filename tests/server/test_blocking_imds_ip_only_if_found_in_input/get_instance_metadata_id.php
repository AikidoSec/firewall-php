<?php

function get_instance_metadata_id($url) {
    $url = 'http://' . $url . '/tests/latest/meta-data/instance-id';

    $ch = curl_init($url);
    curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
    curl_setopt($ch, CURLOPT_HTTPHEADER, ['X-aws-ec2-metadata-token: ' . 'test-token']);
    curl_setopt($ch, CURLOPT_TIMEOUT_MS, 5);

    curl_exec($ch);

    return "test_instance_id";
}
