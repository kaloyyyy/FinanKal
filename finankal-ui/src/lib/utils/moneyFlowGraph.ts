export interface Flow {
    fromAccountId: string;
    toAccountId: string;
    amount: number;
    transactions: number;
}

export interface FlowAccount {
    id: string;
    name: string;
    type: string;
}

export interface GraphNode {
    id: string;
    name: string;
    type: string;
    column: 0 | 1 | 2; // where money comes from | your accounts | where money goes
    x: number;
    y: number;
    r: number;
    inflow: number;
    outflow: number;
}

export interface GraphEdge {
    id: string;
    from: GraphNode;
    to: GraphNode;
    amount: number;
    transactions: number;
    width: number;
    path: string; // SVG path, ending at the arrowhead's base
    arrow: string; // SVG polygon points of the arrowhead
}

const R_MIN = 14;
const R_MAX = 58;
const W_MIN = 1.5;
const W_MAX = 22;
const GAP = 52; // room under / beside a node for its label

/** Income (and equity) on the left, bank and wallet accounts in the middle, spending on the right. */
const columnOf = (type: string): 0 | 1 | 2 =>
    type === "INCOME" || type === "EQUITY" ? 0 : type === "EXPENSE" ? 2 : 1;

/**
 * Lays out who sent money where. A node's area grows with the money passing through it (the larger
 * of what came in and what went out); an arrow's width grows with the money it carries.
 */
export function moneyFlowGraph(accounts: FlowAccount[], flows: Flow[], width = 1000) {
    const byId = new Map(accounts.map((a) => [a.id, a]));
    const nodes = new Map<string, GraphNode>();
    const node = (id: string) => {
        const account = byId.get(id);
        if (!account) return null;
        let n = nodes.get(id);
        if (!n) {
            n = { id, name: account.name, type: account.type, column: columnOf(account.type), x: 0, y: 0, r: R_MIN, inflow: 0, outflow: 0 };
            nodes.set(id, n);
        }
        return n;
    };

    const kept = flows.filter((f) => f.amount > 0 && f.fromAccountId !== f.toAccountId && node(f.fromAccountId) && node(f.toAccountId));
    for (const f of kept) {
        nodes.get(f.fromAccountId)!.outflow += f.amount;
        nodes.get(f.toAccountId)!.inflow += f.amount;
    }

    const through = (n: GraphNode) => Math.max(n.inflow, n.outflow);
    const maxThrough = Math.max(1, ...[...nodes.values()].map(through));
    for (const n of nodes.values()) {
        n.r = R_MIN + (R_MAX - R_MIN) * Math.sqrt(through(n) / maxThrough); // area ∝ money
    }

    // Stack each column biggest first, then centre the shorter columns.
    const xs = [width * 0.15, width * 0.5, width * 0.85];
    const columns = [0, 1, 2].map((c) =>
        [...nodes.values()].filter((n) => n.column === c).sort((a, b) => through(b) - through(a))
    );
    const heights = columns.map((col) => col.reduce((h, n) => h + 2 * n.r + GAP, 0));
    const height = Math.max(200, ...heights) + 20;
    columns.forEach((col, c) => {
        let y = (height - heights[c]) / 2 + 10;
        for (const n of col) {
            n.x = xs[c];
            n.y = y + n.r;
            y += 2 * n.r + GAP;
        }
    });

    const maxAmount = Math.max(1, ...kept.map((f) => f.amount));
    const edges: GraphEdge[] = kept.map((f) => {
        const from = nodes.get(f.fromAccountId)!;
        const to = nodes.get(f.toAccountId)!;
        const w = W_MIN + (W_MAX - W_MIN) * Math.sqrt(f.amount / maxAmount);

        // Between columns: an S-curve. Within a column (transfers between your accounts): an arc out to the
        // side, downward arrows on the right and upward on the left so a two-way pair doesn't overlap.
        let c1: [number, number];
        let c2: [number, number];
        if (from.column !== to.column) {
            const dx = (to.x - from.x) / 2;
            c1 = [from.x + dx, from.y];
            c2 = [to.x - dx, to.y];
        } else {
            const side = from.y < to.y ? 1 : -1;
            const bulge = 80 + Math.abs(to.y - from.y) * 0.4;
            c1 = [from.x + side * bulge, from.y];
            c2 = [to.x + side * bulge, to.y];
        }

        const start = toward([from.x, from.y], c1, from.r);
        const tip = toward([to.x, to.y], c2, to.r + 2);
        const head = 6 + w * 0.9;
        const base = toward([to.x, to.y], c2, to.r + 2 + head);
        const [ux, uy] = unit([tip[0] - base[0], tip[1] - base[1]]);
        const half = w / 2 + 4 + w * 0.25;
        const arrow = [tip, [base[0] - uy * half, base[1] + ux * half], [base[0] + uy * half, base[1] - ux * half]]
            .map(([x, y]) => `${x.toFixed(1)},${y.toFixed(1)}`)
            .join(" ");

        return {
            id: `${f.fromAccountId}>${f.toAccountId}`,
            from,
            to,
            amount: f.amount,
            transactions: f.transactions,
            width: w,
            path: `M${fmt(start)} C${fmt(c1)} ${fmt(c2)} ${fmt(base)}`,
            arrow
        };
    });
    // Thick arrows first, so thin ones stay visible on top.
    edges.sort((a, b) => b.amount - a.amount);

    return { nodes: [...nodes.values()], edges, width, height };
}

function unit([x, y]: [number, number]): [number, number] {
    const length = Math.hypot(x, y) || 1;
    return [x / length, y / length];
}

/** The point `distance` from `center` in the direction of `toward`. */
function toward(center: [number, number], to: [number, number], distance: number): [number, number] {
    const [ux, uy] = unit([to[0] - center[0], to[1] - center[1]]);
    return [center[0] + ux * distance, center[1] + uy * distance];
}

const fmt = ([x, y]: [number, number]) => `${x.toFixed(1)},${y.toFixed(1)}`;
