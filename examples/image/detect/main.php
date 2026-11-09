<?php

declare(strict_types=1);

require_once dirname(__DIR__, 3) . '/vendor/autoload.php';

use Termstrap\Image\Detect;

$proto = Detect::detectProtocol();
echo "Detected terminal graphics protocol: " . $proto->value . "\n";
