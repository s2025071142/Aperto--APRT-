# Aperto (APTO)

Bitcoin’s 21M supply cap establishes it as "digital gold," but its real-world transactional utility remains throttled by layer-1 throughput constraints, rigid smart contract limitations, and long-term state bloat. For long-term viability, it requires structural protocol upgrades—such as fair bell-curve issuance with tail emissions for security subsidies, advanced Turing-complete programmability, and stateless lightweight node synchronisation.

Aperto is a 100% fair-launch Layer-1 blockchain with **no IPO, no ICO, no pre-mine, and no insider allocations.** All coins enter circulation strictly through open, decentralised CPU mining.

---

## Technical Features Implemented In This Code

- **Monero Merge-Mined RandomX PoW Engine (`pow.go`):** Implements Monero's native CPU-friendly mining algorithm with merge-mining support to ensure ASIC resistance while inheriting Monero's established hashrate against botnets and 51% attacks.
- **20-Year Bell-Curve Issuance & Continuous Tail Emission (`mint.go`):** Features a smooth, upside-down U-shaped emission curve peaking over a 20-year horizon—preventing early-miner monopolies—before decaying into a perpetual floor of **0.005 coins** per block forever.
- **Adaptive Dynamic Block Sizing (`abci.go`):** Automatically resizes block execution limits based on transaction traffic trends using a Monero-style block-reward penalty to maintain ultra-low fees under a penny while preventing state bloat.
- **EVM Execution Layer (`go.mod`):** Native compatibility for Solidity smart contracts, decentralised applications (dApps), and standard Web3 wallets like MetaMask.
- **Stateless Light Verification & Dual-Mode Storage (`state.go`):** Separates lightweight wallet verification from full EVM state execution. Supports interchangeable **UTXO Commitment Mode** (stateless cryptographic accumulators on mobile/light devices) and **Pruned Node Mode** (full EVM execution with historical pruning on desktop/PC).

---

## Node Storage & Sync Modes

Node operators can freely switch between stateless UTXO commitments and historical block pruning regardless of their device:

| Node Mode | Default Platform | Storage Footprint | How It Works |
| :--- | :--- | :--- | :--- |
| **UTXO Commitment Mode (`utxo`)** | **Mobile / Light Devices** *(Default)* | **~Kilobytes (KB)** | Designed for stateless light/wallet nodes. Verifies active balances instantly using cryptographic accumulators without downloading historical blocks or storing the full EVM state set. |
| **Pruned Node Mode (`pruned`)** | **Desktop / PC** *(Default)* | **~10–30 GB** | Executes the full EVM state, downloads and validates new blocks in real time, generates proofs for light nodes, and automatically deletes old spent transaction history. |
| **Archival Mode (`archival`)** | **Servers / Explorers** *(Optional)* | **Full History** | Retains every block and historical EVM state transition from Genesis. |

---

## Fair Launch & Tokenomics

- **Zero Pre-Mine / Zero IPO:** 100% of APTO is distributed via public Proof-of-Work block rewards with no VC or founder allocations.
- **Anti-Instamine Bell Curve:** Issuance ramps up gradually over 20 years alongside global adoption before tapering down to the permanent `0.005 APTO` tail emission.
- **Organic Grassroots Liquidity:** Designed for natural, community-driven price discovery without artificial unlock cliffs or insider sell pressure.

---

## Build and Launch Guide

Ensure you have Go 1.21+ installed on your operating system, then execute these commands in your terminal:

```bash
# 1. Download project dependencies and compile the blockchain binary
go mod tidy
go build -o mycustoml1d main.go

# 2. Initialize genesis records and operational configuration states
./mycustoml1d init local-node-test --chain-id l1-custom-network

# 3. Fire up your mining validator node (auto-selects 'utxo' on mobile or 'pruned' on PC)
./mycustoml1d start

# Optional: Explicitly override the default node storage mode on any device
./mycustoml1d start --storage-mode=utxo
./mycustoml1d start --storage-mode=pruned --prune-keep-recent=1000
```

## License
Distributed freely under the open-source MIT License terms.
