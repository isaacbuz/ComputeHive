// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.19;

import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/security/ReentrancyGuard.sol";
import "@openzeppelin/contracts/security/Pausable.sol";
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";

/**
 * @title ComputeEscrow
 * @notice Manages escrow payments for compute jobs with dispute resolution
 * @dev Supports both native ETH and ERC20 token payments
 */
contract ComputeEscrow is Ownable, ReentrancyGuard, Pausable {
    using SafeERC20 for IERC20;

    // Job states
    enum JobStatus {
        Created,
        Funded,
        InProgress,
        Completed,
        Disputed,
        Resolved,
        Cancelled
    }

    // Dispute resolution outcomes
    enum DisputeOutcome {
        None,
        FavorProvider,
        FavorConsumer,
        Split
    }

    struct Job {
        address consumer;
        address provider;
        uint256 amount;
        address paymentToken; // address(0) for ETH
        JobStatus status;
        uint256 createdAt;
        uint256 startedAt;
        uint256 completedAt;
        bytes32 jobHash; // Hash of job details stored off-chain
        DisputeOutcome disputeOutcome;
        uint256 disputeSplitPercentage; // Percentage to provider if split
    }

    struct DisputeInfo {
        uint256 raisedAt;
        address raisedBy;
        string reason;
        uint256 resolvedAt;
        address resolver;
    }

    // State variables
    mapping(uint256 => Job) public jobs;
    mapping(uint256 => DisputeInfo) public disputes;
    mapping(address => uint256[]) public consumerJobs;
    mapping(address => uint256[]) public providerJobs;
    mapping(address => bool) public arbitrators;
    
    uint256 public nextJobId;
    uint256 public platformFeePercentage = 250; // 2.5% in basis points
    uint256 public disputeWindow = 3 days;
    uint256 public minJobDuration = 1 minutes;
    uint256 public totalFeesCollected;
    
    address public feeRecipient;
    
    // Events
    event JobCreated(
        uint256 indexed jobId,
        address indexed consumer,
        address indexed provider,
        uint256 amount,
        address paymentToken
    );
    
    event JobFunded(uint256 indexed jobId);
    event JobStarted(uint256 indexed jobId);
    event JobCompleted(uint256 indexed jobId);
    event JobCancelled(uint256 indexed jobId);
    
    event PaymentReleased(
        uint256 indexed jobId,
        address indexed recipient,
        uint256 amount
    );
    
    event DisputeRaised(
        uint256 indexed jobId,
        address indexed raisedBy,
        string reason
    );
    
    event DisputeResolved(
        uint256 indexed jobId,
        DisputeOutcome outcome,
        address indexed resolver
    );
    
    event ArbitratorAdded(address indexed arbitrator);
    event ArbitratorRemoved(address indexed arbitrator);
    event FeeUpdated(uint256 newFeePercentage);
    
    // Modifiers
    modifier onlyConsumer(uint256 jobId) {
        require(jobs[jobId].consumer == msg.sender, "Not job consumer");
        _;
    }
    
    modifier onlyProvider(uint256 jobId) {
        require(jobs[jobId].provider == msg.sender, "Not job provider");
        _;
    }
    
    modifier onlyParticipant(uint256 jobId) {
        require(
            jobs[jobId].consumer == msg.sender || 
            jobs[jobId].provider == msg.sender,
            "Not a participant"
        );
        _;
    }
    
    modifier onlyArbitrator() {
        require(arbitrators[msg.sender], "Not an arbitrator");
        _;
    }
    
    modifier jobExists(uint256 jobId) {
        require(jobs[jobId].createdAt > 0, "Job does not exist");
        _;
    }
    
    constructor(address _feeRecipient) {
        require(_feeRecipient != address(0), "Invalid fee recipient");
        feeRecipient = _feeRecipient;
        arbitrators[msg.sender] = true;
    }
    
    /**
     * @notice Create a new compute job
     * @param provider Address of the compute provider
     * @param amount Payment amount
     * @param paymentToken Token address (use address(0) for ETH)
     * @param jobHash Hash of job details stored off-chain
     */
    function createJob(
        address provider,
        uint256 amount,
        address paymentToken,
        bytes32 jobHash
    ) external payable whenNotPaused returns (uint256 jobId) {
        require(provider != address(0), "Invalid provider");
        require(provider != msg.sender, "Cannot be own provider");
        require(amount > 0, "Amount must be positive");
        
        jobId = nextJobId++;
        
        Job storage job = jobs[jobId];
        job.consumer = msg.sender;
        job.provider = provider;
        job.amount = amount;
        job.paymentToken = paymentToken;
        job.status = JobStatus.Created;
        job.createdAt = block.timestamp;
        job.jobHash = jobHash;
        
        consumerJobs[msg.sender].push(jobId);
        providerJobs[provider].push(jobId);
        
        emit JobCreated(jobId, msg.sender, provider, amount, paymentToken);
        
        // Auto-fund if payment included
        if (paymentToken == address(0) && msg.value > 0) {
            fundJob(jobId);
        }
    }
    
    /**
     * @notice Fund a created job
     * @param jobId The job ID to fund
     */
    function fundJob(uint256 jobId) 
        public 
        payable 
        whenNotPaused 
        jobExists(jobId) 
        onlyConsumer(jobId) 
    {
        Job storage job = jobs[jobId];
        require(job.status == JobStatus.Created, "Job already funded");
        
        if (job.paymentToken == address(0)) {
            // ETH payment
            require(msg.value >= job.amount, "Insufficient ETH sent");
            if (msg.value > job.amount) {
                // Refund excess
                payable(msg.sender).transfer(msg.value - job.amount);
            }
        } else {
            // ERC20 payment
            IERC20(job.paymentToken).safeTransferFrom(
                msg.sender,
                address(this),
                job.amount
            );
        }
        
        job.status = JobStatus.Funded;
        emit JobFunded(jobId);
    }
    
    /**
     * @notice Provider starts working on the job
     * @param jobId The job ID to start
     */
    function startJob(uint256 jobId) 
        external 
        whenNotPaused 
        jobExists(jobId) 
        onlyProvider(jobId) 
    {
        Job storage job = jobs[jobId];
        require(job.status == JobStatus.Funded, "Job not funded");
        
        job.status = JobStatus.InProgress;
        job.startedAt = block.timestamp;
        
        emit JobStarted(jobId);
    }
    
    /**
     * @notice Provider marks job as completed
     * @param jobId The job ID to complete
     */
    function completeJob(uint256 jobId) 
        external 
        whenNotPaused 
        jobExists(jobId) 
        onlyProvider(jobId) 
    {
        Job storage job = jobs[jobId];
        require(job.status == JobStatus.InProgress, "Job not in progress");
        require(
            block.timestamp >= job.startedAt + minJobDuration,
            "Job duration too short"
        );
        
        job.status = JobStatus.Completed;
        job.completedAt = block.timestamp;
        
        emit JobCompleted(jobId);
    }
    
    /**
     * @notice Release payment to provider after dispute window
     * @param jobId The job ID to release payment for
     */
    function releasePayment(uint256 jobId) 
        external 
        nonReentrant 
        whenNotPaused 
        jobExists(jobId) 
    {
        Job storage job = jobs[jobId];
        require(
            job.status == JobStatus.Completed || 
            (job.status == JobStatus.Resolved && job.disputeOutcome != DisputeOutcome.FavorConsumer),
            "Cannot release payment"
        );
        
        if (job.status == JobStatus.Completed) {
            require(
                block.timestamp >= job.completedAt + disputeWindow,
                "Dispute window not passed"
            );
        }
        
        uint256 amount = job.amount;
        uint256 platformFee = (amount * platformFeePercentage) / 10000;
        uint256 providerAmount = amount - platformFee;
        
        // Handle dispute outcomes
        if (job.disputeOutcome == DisputeOutcome.Split) {
            uint256 providerShare = (amount * job.disputeSplitPercentage) / 100;
            uint256 consumerShare = amount - providerShare;
            platformFee = (providerShare * platformFeePercentage) / 10000;
            providerAmount = providerShare - platformFee;
            
            // Refund consumer their share
            _transferPayment(job.paymentToken, job.consumer, consumerShare);
        }
        
        // Transfer platform fee
        totalFeesCollected += platformFee;
        _transferPayment(job.paymentToken, feeRecipient, platformFee);
        
        // Transfer provider payment
        _transferPayment(job.paymentToken, job.provider, providerAmount);
        
        // Update job status
        job.status = JobStatus.Resolved;
        
        emit PaymentReleased(jobId, job.provider, providerAmount);
    }
    
    /**
     * @notice Cancel a job (only before it starts)
     * @param jobId The job ID to cancel
     */
    function cancelJob(uint256 jobId) 
        external 
        nonReentrant 
        whenNotPaused 
        jobExists(jobId) 
        onlyConsumer(jobId) 
    {
        Job storage job = jobs[jobId];
        require(
            job.status == JobStatus.Created || job.status == JobStatus.Funded,
            "Cannot cancel started job"
        );
        
        if (job.status == JobStatus.Funded) {
            // Refund consumer
            _transferPayment(job.paymentToken, job.consumer, job.amount);
        }
        
        job.status = JobStatus.Cancelled;
        emit JobCancelled(jobId);
    }
    
    /**
     * @notice Raise a dispute for a job
     * @param jobId The job ID to dispute
     * @param reason Reason for the dispute
     */
    function raiseDispute(uint256 jobId, string calldata reason) 
        external 
        whenNotPaused 
        jobExists(jobId) 
        onlyParticipant(jobId) 
    {
        Job storage job = jobs[jobId];
        require(
            job.status == JobStatus.InProgress || job.status == JobStatus.Completed,
            "Cannot dispute job"
        );
        
        if (job.status == JobStatus.Completed) {
            require(
                block.timestamp <= job.completedAt + disputeWindow,
                "Dispute window passed"
            );
        }
        
        job.status = JobStatus.Disputed;
        
        DisputeInfo storage dispute = disputes[jobId];
        dispute.raisedAt = block.timestamp;
        dispute.raisedBy = msg.sender;
        dispute.reason = reason;
        
        emit DisputeRaised(jobId, msg.sender, reason);
    }
    
    /**
     * @notice Resolve a dispute (arbitrator only)
     * @param jobId The job ID with dispute
     * @param outcome The dispute resolution outcome
     * @param splitPercentage Percentage to provider if split outcome
     */
    function resolveDispute(
        uint256 jobId,
        DisputeOutcome outcome,
        uint256 splitPercentage
    ) external whenNotPaused jobExists(jobId) onlyArbitrator {
        Job storage job = jobs[jobId];
        require(job.status == JobStatus.Disputed, "No dispute to resolve");
        
        if (outcome == DisputeOutcome.Split) {
            require(
                splitPercentage > 0 && splitPercentage < 100,
                "Invalid split percentage"
            );
            job.disputeSplitPercentage = splitPercentage;
        }
        
        job.disputeOutcome = outcome;
        job.status = JobStatus.Resolved;
        
        DisputeInfo storage dispute = disputes[jobId];
        dispute.resolvedAt = block.timestamp;
        dispute.resolver = msg.sender;
        
        emit DisputeResolved(jobId, outcome, msg.sender);
        
        // Auto-release payment based on outcome
        if (outcome == DisputeOutcome.FavorConsumer) {
            // Refund consumer
            _transferPayment(job.paymentToken, job.consumer, job.amount);
        } else {
            // FavorProvider or Split - handled in releasePayment
            releasePayment(jobId);
        }
    }
    
    /**
     * @notice Add an arbitrator
     * @param arbitrator Address to add as arbitrator
     */
    function addArbitrator(address arbitrator) external onlyOwner {
        require(arbitrator != address(0), "Invalid arbitrator");
        require(!arbitrators[arbitrator], "Already an arbitrator");
        
        arbitrators[arbitrator] = true;
        emit ArbitratorAdded(arbitrator);
    }
    
    /**
     * @notice Remove an arbitrator
     * @param arbitrator Address to remove as arbitrator
     */
    function removeArbitrator(address arbitrator) external onlyOwner {
        require(arbitrators[arbitrator], "Not an arbitrator");
        
        arbitrators[arbitrator] = false;
        emit ArbitratorRemoved(arbitrator);
    }
    
    /**
     * @notice Update platform fee percentage
     * @param newFeePercentage New fee in basis points (e.g., 250 = 2.5%)
     */
    function updatePlatformFee(uint256 newFeePercentage) external onlyOwner {
        require(newFeePercentage <= 1000, "Fee too high"); // Max 10%
        
        platformFeePercentage = newFeePercentage;
        emit FeeUpdated(newFeePercentage);
    }
    
    /**
     * @notice Update fee recipient address
     * @param newRecipient New fee recipient address
     */
    function updateFeeRecipient(address newRecipient) external onlyOwner {
        require(newRecipient != address(0), "Invalid recipient");
        feeRecipient = newRecipient;
    }
    
    /**
     * @notice Pause the contract
     */
    function pause() external onlyOwner {
        _pause();
    }
    
    /**
     * @notice Unpause the contract
     */
    function unpause() external onlyOwner {
        _unpause();
    }
    
    /**
     * @notice Withdraw accumulated fees
     */
    function withdrawFees() external onlyOwner {
        uint256 balance = address(this).balance;
        require(balance > 0, "No fees to withdraw");
        
        payable(feeRecipient).transfer(balance);
    }
    
    // Internal functions
    function _transferPayment(
        address token,
        address recipient,
        uint256 amount
    ) private {
        if (token == address(0)) {
            // ETH transfer
            payable(recipient).transfer(amount);
        } else {
            // ERC20 transfer
            IERC20(token).safeTransfer(recipient, amount);
        }
    }
    
    // View functions
    function getJob(uint256 jobId) external view returns (Job memory) {
        return jobs[jobId];
    }
    
    function getConsumerJobs(address consumer) external view returns (uint256[] memory) {
        return consumerJobs[consumer];
    }
    
    function getProviderJobs(address provider) external view returns (uint256[] memory) {
        return providerJobs[provider];
    }
    
    function getDisputeInfo(uint256 jobId) external view returns (DisputeInfo memory) {
        return disputes[jobId];
    }
    
    function canReleasePayment(uint256 jobId) external view returns (bool) {
        Job memory job = jobs[jobId];
        
        if (job.status == JobStatus.Completed) {
            return block.timestamp >= job.completedAt + disputeWindow;
        }
        
        if (job.status == JobStatus.Resolved) {
            return job.disputeOutcome != DisputeOutcome.FavorConsumer;
        }
        
        return false;
    }
} 