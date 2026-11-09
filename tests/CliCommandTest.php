<?php

declare(strict_types=1);

namespace Termstrap\Tests;

use PHPUnit\Framework\TestCase;
use Symfony\Component\Console\Application;
use Symfony\Component\Console\Tester\CommandTester;
use Termstrap\Command\RenderCommand;

final class CliCommandTest extends TestCase
{
    public function testRenderCommandExecutesOnRawHtml(): void
    {
        $app = new Application();
        $cmd = new RenderCommand();
        $app->addCommand($cmd);

        $tester = new CommandTester($cmd);
        $tester->execute([
            'input' => '<h1>Hello CLI</h1>',
            '--width' => 40,
            '--theme' => 'none',
        ]);

        $output = $tester->getDisplay();
        $this->assertSame(0, $tester->getStatusCode());
        $this->assertStringContainsString('Hello CLI', $output);
    }
}
