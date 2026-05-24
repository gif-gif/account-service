import { ButtonHTMLAttributes, forwardRef } from "react";

import { cn } from "../../lib/cn";

type ButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: "default" | "secondary" | "ghost" | "destructive";
  size?: "default" | "sm" | "icon";
};

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant = "default", size = "default", ...props }, ref) => (
    <button ref={ref} className={cn("ui-button", `ui-button--${variant}`, `ui-button--${size}`, className)} {...props} />
  ),
);

Button.displayName = "Button";
