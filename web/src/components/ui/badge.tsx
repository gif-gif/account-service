import { HTMLAttributes, forwardRef } from "react";

import { cn } from "../../lib/cn";

type BadgeProps = HTMLAttributes<HTMLSpanElement> & {
  variant?: "default" | "secondary" | "success" | "warning" | "destructive";
};

export const Badge = forwardRef<HTMLSpanElement, BadgeProps>(({ className, variant = "default", ...props }, ref) => (
  <span ref={ref} className={cn("ui-badge", `ui-badge--${variant}`, className)} {...props} />
));

Badge.displayName = "Badge";
