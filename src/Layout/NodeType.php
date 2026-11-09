<?php

declare(strict_types=1);

namespace Termstrap\Layout;

enum NodeType
{
    case Element;
    case Text;
    case Image;
}
