<?php
// Synthetic fixture only. No environment configuration or production endpoints.
define('ENVIRONMENT', 'testing');
define('FCPATH', __DIR__.'/');
define('SELF', 'index.php');
define('BASEPATH', __DIR__.'/vendor/codeigniter/framework/system/');
define('APPPATH', __DIR__.'/application/');
define('VIEWPATH', APPPATH.'views/');
require BASEPATH.'core/CodeIgniter.php';
