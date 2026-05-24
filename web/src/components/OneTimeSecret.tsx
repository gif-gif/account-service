import { useState } from "react";

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
    <section>
      <strong>{label}</strong>
      <code>{value}</code>
      <button type="button" onClick={() => setVisible(false)}>
        Dismiss {label}
      </button>
    </section>
  );
}
