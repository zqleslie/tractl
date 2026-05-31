export interface IndependentDividerProps {
  x: number
  height: number
}

export function IndependentDivider({ x, height }: IndependentDividerProps) {
  return (
    <div className="pointer-events-none absolute" style={{ left: x, top: 0, height }}>
      <svg width={1} height={height} className="block overflow-visible">
        <line
          x1={0}
          y1={0}
          x2={0}
          y2={height}
          stroke="#CBD5E1"
          strokeWidth={0.5}
          strokeDasharray="4 4"
        />
      </svg>
      <span className="absolute left-0 top-1/2 -translate-x-1/2 -translate-y-1/2 rotate-90 whitespace-nowrap bg-white px-1 text-[10px] font-[400] text-slate-400 dark:bg-[#16171d] dark:text-slate-500">
        independent
      </span>
    </div>
  )
}
