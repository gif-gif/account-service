import { useState } from "react";

import { Button } from "./ui/button";

type Props = {
  label: string;
  value: string;
};

export function OneTimeSecret({ label, value }: Props) {
  const [visible, setVisible] = useState(true);

  if (!visible) {
    return null;
  }

  return (
    <section className="secret-box">
      <strong>{label}</strong>
      <code>{value}</code>
      <Button size="sm" type="button" variant="secondary" onClick={() => setVisible(false)}>
        Dismiss {label}
      </Button>
    </section>
  );
}
