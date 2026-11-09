<?php

declare(strict_types=1);

$root = dirname(__DIR__, 2);
$oracleFile = $root . '/termstrap.ffi/Termstrap.php';

if (!file_exists($oracleFile)) {
    fwrite(STDERR, "Oracle file not found: {$oracleFile}\n");
    exit(1);
}

require_once $oracleFile;

$input = file_get_contents('php://stdin');
if ($input === false || $input === '') {
    echo json_encode(['error' => 'No input provided'], JSON_THROW_ON_ERROR);
    exit(0);
}

try {
    $data = json_decode($input, true, 512, JSON_THROW_ON_ERROR);
    $content = $data['content'] ?? '';
    $options = $data['options'] ?? [];

    $termstrap = new Termstrap\Termstrap();
    $output = $termstrap->render($content, $options);

    echo json_encode(['output' => $output], JSON_THROW_ON_ERROR);
} catch (Throwable $e) {
    echo json_encode(['error' => $e->getMessage()], JSON_THROW_ON_ERROR);
}
