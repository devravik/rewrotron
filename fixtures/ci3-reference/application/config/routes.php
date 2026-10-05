<?php
$route['default_controller'] = 'reference/health';
$route['404_override'] = '';
$route['translate_uri_dashes'] = false;
foreach (['health', 'login', 'whoami', 'csrf', 'create', 'item', 'note', 'logout'] as $name) {
    $route[$name] = 'reference/'.$name;
}
