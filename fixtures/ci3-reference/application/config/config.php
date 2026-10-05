<?php
// Closed, synthetic fixture configuration; deliberately no getenv/.env loader.
$config = [
    'base_url' => 'http://source:8080/', 'index_page' => '',
    'uri_protocol' => 'REQUEST_URI', 'url_suffix' => '',
    'language' => 'english', 'charset' => 'UTF-8', 'enable_hooks' => false,
    'subclass_prefix' => 'MY_', 'composer_autoload' => false,
    'permitted_uri_chars' => 'a-z 0-9~%.:_\\-', 'enable_query_strings' => false,
    'controller_trigger' => 'c', 'function_trigger' => 'm', 'directory_trigger' => 'd',
    'allow_get_array' => true, 'log_threshold' => 0, 'log_path' => '/tmp/',
    'log_file_extension' => '', 'log_file_permissions' => 0644,
    'log_date_format' => 'Y-m-d H:i:s', 'error_views_path' => '/app/vendor/codeigniter/framework/application/views/errors/',
    'cache_path' => '', 'cache_query_string' => false, 'encryption_key' => '',
    'cookie_prefix' => '', 'cookie_domain' => '', 'cookie_path' => '/',
    'cookie_secure' => false, 'cookie_httponly' => true,
    'standardize_newlines' => false, 'global_xss_filtering' => false,
    // CSRF/session binding belongs to the explicit JSON session driver below.
    'csrf_protection' => false, 'csrf_token_name' => 'csrf',
    'csrf_cookie_name' => 'csrf', 'csrf_expire' => 30, 'csrf_regenerate' => true,
    'csrf_exclude_uris' => [], 'compress_output' => false,
    'time_reference' => 'gmt', 'rewrite_short_tags' => false, 'proxy_ips' => ''
];
