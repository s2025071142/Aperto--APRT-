# Aperto (APTO)

Bitcoin’s 21M supply cap establishes it as "digital gold," but its real-world transactional utility remains throttled by layer-1 throughput constraints, rigid smart contract limitations, and long-term state bloat. For long-term viability, it requires structural protocol upgrades—such as fair bell-curve issuance with perpetual tail emissions, fast sub-2-second block lock-in, advanced Turing-complete programmability, and stateless lightweight node synchronization.

Aperto is a 100% fair-launch Layer-1 blockchain with **no IPO, no ICO, no pre-mine, and no insider allocations.** All coins enter circulation strictly through open, decentralized CPU mining.

---

## Technical Features Implemented In This Code

- **Monero Merge-Mined RandomX PoW & Fast Pre-Sign Lock-In (`pow.go`):** Implements Monero's native CPU-friendly mining algorithm with merge-mining support to ensure ASIC resistance while inheriting Monero's established hashrate against botnets and 51% attacks. Includes sub-2-second pre-sign block verification so nodes lock in valid blocks rapidly and immediately halt redundant hashing on the same height.
- **21M Bell-Curve Issuance & Continuous Tail Emission (`mint.go`):** Runs on a **1-minute block time** with a smooth, upside-down U-shaped emission curve that distributes **21,000,000 APTO** over 20 years (peaking at ~2.992 APTO/block at Year 10)—preventing early-miner monopolies—before locking into a perpetual floor of **0.005 APTO** per block forever.
- **Adaptive Dynamic Block Sizing (`abci.go`):** Automatically resizes block execution limits based on transaction traffic trends using a Monero-style block-reward penalty to maintain ultra-low fees under a penny while preventing state bloat.
- **EVM Execution Layer (`go.mod`):** Native compatibility for Solidity smart contracts, decentralized applications (dApps), and standard Web3 wallets like MetaMask.
- **Stateless Light Verification & Dual-Mode Storage (`state.go`):** Separates lightweight wallet verification from full EVM state execution. Supports interchangeable **UTXO Commitment Mode** (stateless cryptographic accumulators on mobile/light devices) and **Pruned Node Mode** (full EVM execution with historical pruning on desktop/PC).

---

## Tokenomics & Network Parameters

| Parameter | Specification |
| :--- | :--- |
| **Launch Type** | 100% Fair Launch (Zero Pre-Mine, Zero IPO/ICO, Zero Insider Allocation) |
| **Target Block Time** | **1 Minute** (~525,960 blocks/year) |
| **Block Lock-In Latency** | **< 2 Seconds** (Pre-sign rapid validation to prevent wasted CPU cycles) |
| **Primary Supply (Years 1–20)** | **21,000,000 APTO** distributed via a 20-year upside-down U bell curve |
| **Peak Block Reward (Year 10)** | **~2.992 APTO** per block (at Block 5,259,600) |
| **Tail Emission (Year 20+)** | **0.005 APTO** per block forever (permanent security subsidy) |

---

## Node Storage & Sync Modes

Node operators can freely switch between stateless UTXO commitments and historical block pruning regardless of their device:

| Node Mode | Default Platform | Storage Footprint | How It Works |
| :--- | :--- | :--- | :--- |
| **UTXO Commitment Mode (`utxo`)** | **Mobile / Light Devices** *(Default)* | **~Kilobytes (KB)** | Designed for stateless light/wallet nodes. Verifies active balances instantly using cryptographic accumulators without downloading historical blocks or storing the full EVM state set. |
| **Pruned Node Mode (`pruned`)** | **Desktop / PC** *(Default)* | **~10–30 GB** | Executes the full EVM state, downloads and validates new blocks in real time, generates proofs for light nodes, and automatically deletes old spent transaction history. |
| **Archival Mode (`archival`)** | **Servers / Explorers** *(Optional)* | **Full History** | Retains every block and historical EVM state transition from Genesis. |

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
