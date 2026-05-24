import { useEffect, useState } from "react";

type Props = {
  label: string;
  value: string;
};

export function RevealSecret({ label, value }: Props) {
  const [revealed, setRevealed] = useState(false);

  useEffect(() => () => setRevealed(false), []);

  return (
    <div>
      <span>{label}</span>
      {revealed ? <code>{value}</code> : null}
      <button type="button" onClick={() => setRevealed(true)}>
        Reveal {label}
      </button>
    </div>
  );
}
