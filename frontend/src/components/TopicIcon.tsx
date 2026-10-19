import {
  BrainCircuit, Bot, Plug, Search, Boxes, ClipboardCheck, Rocket,
  type LucideIcon,
} from 'lucide-react';

const MAP: Record<string, LucideIcon> = {
  BrainCircuit, Bot, Plug, Search, Boxes, ClipboardCheck, Rocket,
};

export function TopicIcon({ name, className }: { name: string; className?: string }): JSX.Element {
  const Cmp = MAP[name] ?? BrainCircuit;
  return <Cmp className={className} />;
}
