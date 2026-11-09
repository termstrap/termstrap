<?php

declare(strict_types=1);

namespace Termstrap\Command;

use Symfony\Component\Console\Attribute\AsCommand;
use Symfony\Component\Console\Command\Command;
use Symfony\Component\Console\Input\InputArgument;
use Symfony\Component\Console\Input\InputInterface;
use Symfony\Component\Console\Input\InputOption;
use Symfony\Component\Console\Output\OutputInterface;
use Symfony\Component\Console\Terminal;
use Termstrap\Termstrap;
use Throwable;

#[AsCommand(
    name: 'termstrap:render',
    description: 'Render HTML/CSS/Markdown content directly in the terminal'
)]
final class RenderCommand extends Command
{
    protected function configure(): void
    {
        $this
            ->addArgument('input', InputArgument::OPTIONAL, 'File path, HTML/Markdown content, or "-" for stdin')
            ->addOption('width', 'w', InputOption::VALUE_REQUIRED, 'Terminal width in columns (0 = auto-detect)', 0)
            ->addOption('theme', 't', InputOption::VALUE_REQUIRED, 'Visual theme (bootstrap, tokyonight, dracula, glamour, none)', 'bootstrap')
            ->addOption('root', 'r', InputOption::VALUE_REQUIRED, 'Root directory for resolving local relative image paths', '')
            ->addOption('color', 'c', InputOption::VALUE_REQUIRED, 'Color mode (truecolor, 256, 16)', '')
            ->addOption('protocol', 'p', InputOption::VALUE_REQUIRED, 'Force image protocol (halfblock, sextant, kitty, iterm, sixel)', '')
            ->addOption('css', null, InputOption::VALUE_REQUIRED, 'Path to custom CSS stylesheet file', '')
            ->addOption('no-images', null, InputOption::VALUE_NONE, 'Disable image rendering and replace with text placeholders')
            ->addOption('optimize', null, InputOption::VALUE_NONE, 'Deduplicate contiguous ANSI escape sequences')
            ->addOption('watch', 'W', InputOption::VALUE_NONE, 'Watch file for changes and automatically reload render')
            ->addOption('link-mode', null, InputOption::VALUE_REQUIRED, 'Hyperlink presentation mode (osc8, inline, footnotes)', '')
            ->addOption('image-mode', null, InputOption::VALUE_REQUIRED, 'Image rendering mode (auto, smart, glamour, placeholder)', '')
            ->addOption('min-image-width', null, InputOption::VALUE_REQUIRED, 'Minimum column width for images before smart fallback', 0)
            ->addOption('preset', null, InputOption::VALUE_REQUIRED, 'Preconfigured presentation preset (e.g. glamour)', '')
            ->addOption('markdown', 'm', InputOption::VALUE_NONE, 'Treat input as Markdown and convert to HTML');
    }

    protected function execute(InputInterface $input, OutputInterface $output): int
    {
        $inputArg = $input->getArgument('input');
        $rawContent = '';
        $inputFilePath = '';
        $rootPath = (string) $input->getOption('root');

        if ($inputArg === null || $inputArg === '-') {
            // Read from stdin
            $rawContent = (string) file_get_contents('php://stdin');
            if ($rootPath === '') {
                $rootPath = '.';
            }
        } elseif (file_exists($inputArg)) {
            $inputFilePath = $inputArg;
            $rawContent = (string) file_get_contents($inputFilePath);
            if ($rootPath === '') {
                $rootPath = dirname($inputFilePath);
            }
        } elseif (str_contains($inputArg, '<') && str_contains($inputArg, '>')) {
            // Raw HTML passed directly
            $rawContent = $inputArg;
            if ($rootPath === '') {
                $rootPath = '.';
            }
        } else {
            $output->writeln("<error>Error: input file not found: {$inputArg}</error>");
            return Command::FAILURE;
        }

        $width = (int) $input->getOption('width');
        if ($width <= 0) {
            $terminal = new Terminal();
            $width = $terminal->getWidth();
            if ($width <= 0) {
                $width = 80;
            }
        }

        $theme = (string) $input->getOption('theme');
        $linkMode = (string) $input->getOption('link-mode');
        $imageMode = (string) $input->getOption('image-mode');
        $preset = strtolower((string) $input->getOption('preset'));

        if ($preset === 'glamour') {
            $theme = 'glamour';
            if ($linkMode === '') {
                $linkMode = 'inline';
            }
            if ($imageMode === '') {
                $imageMode = 'glamour';
            }
        }

        // Markdown detection
        $markdownMode = (bool) $input->getOption('markdown');
        if (!$markdownMode && $inputFilePath !== '') {
            $ext = strtolower(pathinfo($inputFilePath, PATHINFO_EXTENSION));
            if ($ext === 'md' || $ext === 'markdown') {
                $markdownMode = true;
            }
        }

        $stylesheets = [];
        $cssFile = (string) $input->getOption('css');
        if ($cssFile !== '') {
            if (!file_exists($cssFile)) {
                $output->writeln("<error>Error: CSS file not found: {$cssFile}</error>");
                return Command::FAILURE;
            }
            $stylesheets[] = (string) file_get_contents($cssFile);
        }

        $protocol = (string) $input->getOption('protocol');
        if ($protocol !== '') {
            putenv('TERMSTRAP_IMAGE_PROTOCOL=' . $protocol);
        }

        $options = [
            'width' => $width,
            'theme' => $theme,
            'root_path' => $rootPath,
            'stylesheets' => $stylesheets,
            'link_mode' => $linkMode,
            'image_mode' => $imageMode,
            'disable_images' => (bool) $input->getOption('no-images'),
            'markdown' => $markdownMode,
            'min_image_width' => (int) $input->getOption('min-image-width'),
            'color_mode' => (string) $input->getOption('color'),
        ];

        try {
            $rendered = Termstrap::toAnsi($rawContent, $options);
            $output->write($rendered);
            return Command::SUCCESS;
        } catch (Throwable $e) {
            $output->writeln('<error>Render error: ' . $e->getMessage() . '</error>');
            return Command::FAILURE;
        }
    }
}
