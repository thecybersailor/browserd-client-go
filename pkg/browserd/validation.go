package browserd

import "strings"

func ValidateActInput(input ActInput) error {
	action := strings.TrimSpace(input.Action)
	if action == "" {
		return Error{Code: "browserd_act_action_required", Message: "action is required"}
	}
	ref := strings.TrimSpace(input.Ref)
	switch action {
	case "click":
		if ref == "" && (input.X <= 0 || input.Y <= 0) {
			return Error{Code: "browserd_act_target_required", Message: "click requires ref or x/y"}
		}
	case "doubleClick", "hover", "scrollIntoView", "select":
		if ref == "" {
			return Error{Code: "browserd_act_ref_required", Message: action + " requires ref"}
		}
	case "scroll":
		if input.DeltaX == 0 && input.DeltaY == 0 {
			return Error{Code: "browserd_act_scroll_delta_required", Message: "scroll requires deltaX or deltaY"}
		}
	case "paste":
		if ref == "" && (input.X <= 0 || input.Y <= 0) {
			return Error{Code: "browserd_act_target_required", Message: "paste requires ref or x/y"}
		}
		if strings.TrimSpace(input.Text) == "" && strings.TrimSpace(input.HTML) == "" {
			return Error{Code: "browserd_act_text_required", Message: "paste requires text or html"}
		}
	case "type":
		if ref == "" || strings.TrimSpace(input.Text) == "" {
			return Error{Code: "browserd_act_type_input_required", Message: "type requires ref and text"}
		}
	case "fill":
		if ref == "" {
			return Error{Code: "browserd_act_ref_required", Message: "fill requires ref"}
		}
	case "press":
		if ref == "" {
			return Error{Code: "browserd_act_ref_required", Message: "press requires ref"}
		}
		if !isAllowedPressKey(input.Key) {
			return Error{Code: "INVALID_KEY", Message: "press key must be a W3C KeyboardEvent.key name or one printable character"}
		}
	default:
		return Error{Code: "browserd_act_action_invalid", Message: "unsupported action: " + action}
	}
	return nil
}

func ValidateScreenshotInput(input ScreenshotInput) error {
	mode := strings.TrimSpace(input.Mode)
	if mode == "" {
		mode = "viewport"
	}
	switch mode {
	case "viewport", "fullpage":
		if strings.TrimSpace(input.Selector) != "" {
			return Error{Code: "browserd_screenshot_selector_mode_required", Message: "selector is only supported when mode is selector"}
		}
	case "selector":
		if strings.TrimSpace(input.Selector) == "" {
			return Error{Code: "browserd_screenshot_selector_required", Message: "selector is required when mode is selector"}
		}
	default:
		return Error{Code: "browserd_screenshot_mode_invalid", Message: "mode must be viewport, fullpage, or selector"}
	}
	return nil
}

func ValidatePageToolInput(input PageToolInput) error {
	method := strings.TrimSpace(input.Method)
	if method == "" {
		return Error{Code: "browserd_pagetool_method_required", Message: "pageTool method is required"}
	}
	if !IsSupportedPageToolMethod(method) {
		return Error{Code: "browserd_pagetool_method_unsupported", Message: "pageTool method is not supported: " + method}
	}
	if input.TimeoutMs < 0 {
		return Error{Code: "browserd_timeout_invalid", Message: "timeoutMs must be greater than or equal to 0"}
	}
	return nil
}

func isAllowedPressKey(key string) bool {
	key = strings.TrimSpace(key)
	if len([]rune(key)) == 1 && key != "" {
		return true
	}
	switch key {
	case "ArrowDown", "ArrowLeft", "ArrowRight", "ArrowUp", "Backspace", "Delete", "End", "Enter", "Escape", "Home", "PageDown", "PageUp", "Space", "Tab":
		return true
	default:
		return false
	}
}
