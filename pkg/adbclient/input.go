package adbclient

import (
	"fmt"
	"strings"
)

// InputSource returns the input source of the device
type InputSource int

const (
	InputSourceDefault InputSource = iota
	InputSourceDpad
	InputSourceKeyboard
	InputSourceMouse
	InputSourceTouchpad
	InputSourceGamepad
	InputSourceTouchNavigation
	InputSourceJoystick
	InputSourceTouchScreen
	InputSourceTouchStylus
	InputSourceTrackball
)

func (s InputSource) String() string {
	switch s {
	case InputSourceDefault:
		return "default"
	case InputSourceDpad:
		return "dpad"
	case InputSourceKeyboard:
		return "keyboard"
	case InputSourceMouse:
		return "mouse"
	case InputSourceTouchpad:
		return "touchpad"
	case InputSourceGamepad:
		return "gamepad"
	case InputSourceTouchNavigation:
		return "touchnavigation"
	case InputSourceJoystick:
		return "joystick"
	case InputSourceTouchScreen:
		return "touchscreen"
	case InputSourceTouchStylus:
		return "touchstylus"
	case InputSourceTrackball:
		return "trackball"
	default:
		return "unknown"
	}
}

type InputCommand int

const (
	InputCommandText InputCommand = iota
	InputCommandKeyEvent
	InputCommandTap
	InputCommandSwipe
	InputCommandDragAndDrop
	InputCommandPress
	InputCommandRoll
	InputCommandMotionEvent
	InputCommandKeyCombination
)

func (c InputCommand) String() string {
	switch c {
	case InputCommandText:
		return "text"
	case InputCommandKeyEvent:
		return "keyevent"
	case InputCommandTap:
		return "tap"
	case InputCommandSwipe:
		return "swipe"
	case InputCommandDragAndDrop:
		return "draganddrop"
	case InputCommandPress:
		return "press"
	case InputCommandRoll:
		return "roll"
	case InputCommandMotionEvent:
		return "motionevent"
	case InputCommandKeyCombination:
		return "keycombination"
	default:
		return "unknown"
	}
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}

	return false
}

func (c InputCommand) ValidateArgs(args ...interface{}) error {
	switch c {
	case InputCommandText:
		if len(args) != 1 {
			return fmt.Errorf("expected 1 argument for %s, got %d", c, len(args))
		}
		if _, ok := args[0].(string); !ok {
			return fmt.Errorf("expected string argument for %s, got %T", c, args[0])
		}

	case InputCommandKeyEvent:
		if len(args) < 1 {
			return fmt.Errorf("expected at least 1 argument for %s, got %d", c, len(args))
		}

		if arg0, ok := args[0].(string); ok {
			if arg0 != "--longpress" && arg0 != "--doubletap" {
				return fmt.Errorf("expected --longpress or --doubletap as first argument for %s, got %s", c, arg0)
			}
		} else {
			if _, ok := args[0].(int); !ok {
				return fmt.Errorf("expected int or string argument as first argument for %s, got %T", c, args[0])
			}
		}

		for i, arg := range args[1:] {
			if _, ok := arg.(int); !ok {
				return fmt.Errorf("expected int argument at index %d for %s, got %T", i+1, c, arg)
			}
		}

	case InputCommandTap:
		if len(args) != 2 {
			return fmt.Errorf("expected 2 arguments for %s, got %d", c, len(args))
		}

		for i, arg := range args {
			if _, ok := arg.(int); !ok {
				return fmt.Errorf("expected int argument at index %d for %s, got %T", i, c, arg)
			}
		}

	case InputCommandSwipe:
		if len(args) < 4 || len(args) > 5 {
			return fmt.Errorf("expected 4 arguments for %s, got %d", c, len(args))
		}

		for i, arg := range args {
			if _, ok := arg.(int); !ok {
				return fmt.Errorf("expected int argument at index %d for %s, got %T", i, c, arg)
			}
		}

	case InputCommandDragAndDrop:
		if len(args) < 4 || len(args) > 5 {
			return fmt.Errorf("expected 4 arguments for %s, got %d", c, len(args))
		}

		for i, arg := range args {
			if _, ok := arg.(int); !ok {
				return fmt.Errorf("expected int argument at index %d for %s, got %T", i, c, arg)
			}
		}

	case InputCommandPress:
		if len(args) != 0 {
			return fmt.Errorf("expected 0 arguments for %s, got %d", c, len(args))
		}

	case InputCommandRoll:
		if len(args) != 2 {
			return fmt.Errorf("expected 2 argument for %s, got %d", c, len(args))
		}

		for i, arg := range args {
			if _, ok := arg.(int); !ok {
				return fmt.Errorf("expected int argument at index %d for %s, got %T", i, c, arg)
			}
		}

	case InputCommandMotionEvent:
		if len(args) != 3 {
			return fmt.Errorf("expected 3 argument for %s, got %d", c, len(args))
		}

		if arg0, ok := args[0].(string); ok {
			var validEvents = []string{"down", "up", "move", "cancel"}
			if !contains(validEvents, arg0) {
				return fmt.Errorf("expected one of %v as first argument for %s, got %s", validEvents, c, arg0)
			}
		} else {
			return fmt.Errorf("expected string argument as first argument for %s, got %T", c, args[0])
		}

		for i, arg := range args[1:] {
			if _, ok := arg.(int); !ok {
				return fmt.Errorf("expected int argument at index %d for %s, got %T", i+1, c, arg)
			}
		}

	case InputCommandKeyCombination:
		if len(args) < 1 {
			return fmt.Errorf("expected at least 1 argument for %s, got %d", c, len(args))
		}

		for i, arg := range args {
			if _, ok := arg.(string); !ok {
				return fmt.Errorf("expected string argument at index %d for %s, got %T", i, c, arg)
			}
		}

	default:
		return fmt.Errorf("unknown command %s", c)
	}

	return nil
}

// Input sends input to the device
func (c *Client) Input(device *Device, source InputSource, command InputCommand, args ...interface{}) error {
	c.log.Infof("Sending input %s %s %v...", source, command, args)

	if err := command.ValidateArgs(args...); err != nil {
		return err
	}

	var a []string
	for _, arg := range args {
		a = append(a, fmt.Sprintf("%v", arg))
	}

	s := source.String()
	if source == InputSourceDefault {
		s = ""
	}

	_, err := c.runCommand(device, "input", fmt.Sprintf("%s %s %s", s, command, strings.Join(a, " ")))
	return err
}
