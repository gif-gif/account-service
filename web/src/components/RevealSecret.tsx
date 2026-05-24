import { useEffect, useState } from "react";

import { Button } from "./ui/button";

type Props = {
  label: string;
  value: string;
};

export function RevealSecret({ label, value }: Props) {
  const [revealed, setRevealed] = useState(false);

  useEffect(() => () => setRevealed(false), []);

  return (
    <div className="secret-box">
      <span>{label}</span>
      {revealed ? <code>{value}</code> : null}
      <Button size="sm" type="button" variant="secondary" onClick={() => setRevealed(true)}>
        Reveal {label}
      </Button>
    </div>
  );
}
