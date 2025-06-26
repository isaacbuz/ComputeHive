<<<<<<< HEAD
// SPDX-License-Identifier: MIT
=======
// SPDX-License-Identifier: Apache-2.0
>>>>>>> 4c40309e804c8f522625b7fd70da67d8d7383849
pragma solidity ^0.8.19;

import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/security/ReentrancyGuard.sol";
import "@openzeppelin/contracts/security/Pausable.sol";
<<<<<<< HEAD
import "@openzeppelin/contracts/utils/math/SafeMath.sol";
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@chainlink/contracts/src/v0.8/interfaces/AggregatorV3Interface.sol";

/**
 * @title ComputeEscrow
 * @dev Main escrow contract for ComputeHive distributed compute jobs
 */
contract ComputeEscrow is Ownable, ReentrancyGuard, Pausable {
    using SafeMath for uint256;
=======
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";

/**
 * @title ComputeEscrow
 * @notice Manages escrow payments for compute jobs with dispute resolution
 * @dev Supports both native ETH and ERC20 token payments
 */
contract ComputeEscrow is Ownable, ReentrancyGuard, Pausable {
    using SafeERC20 for IERC20;
>>>>>>> 4c40309e804c8f522625b7fd70da67d8d7383849

    // Job states
    enum JobStatus {
        Created,
<<<<<<< HEAD
        Assigned,
        Running,
        Completed,
        Disputed,
        Cancelled,
        Failed
    }

    // Job structure
    struct Job {
        address requester;
        address provider;
        uint256 amount;
        uint256 collateral;
        bytes32 jobHash;
        bytes32 resultHash;
        uint256 deadline;
        JobStatus status;
        bool resultVerified;
        uint256 createdAt;
        uint256 completedAt;
=======
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
>>>>>>> 4c40309e804c8f522625b7fd70da67d8d7383849
    }

    // State variables
    mapping(uint256 => Job) public jobs;
<<<<<<< HEAD
    mapping(address => uint256[]) public requesterJobs;
    mapping(address => uint256[]) public providerJobs;
    mapping(address => uint256) public providerReputation;
    mapping(address => uint256) public requesterReputation;
    mapping(address => bool) public verifiers;
    
    uint256 public jobCounter;
    uint256 public platformFeePercent = 250; // 2.5%
    uint256 public collateralPercent = 1000; // 10%
    uint256 public disputeTimeout = 24 hours;
    address public feeRecipient;
    
    IERC20 public paymentToken;
    AggregatorV3Interface public priceFeed;

    // Events
    event JobCreated(
        uint256 indexed jobId,
        address indexed requester,
        uint256 amount,
        bytes32 jobHash,
        uint256 deadline
    );
    
    event JobAssigned(
        uint256 indexed jobId,
        address indexed provider
    );
    
    event JobCompleted(
        uint256 indexed jobId,
        bytes32 resultHash,
        uint256 completedAt
    );
    
    event JobDisputed(
        uint256 indexed jobId,
        address indexed disputer,
        string reason
    );
    
    event JobResolved(
        uint256 indexed jobId,
        bool requesterWins
    );
    
    event PaymentReleased(
        uint256 indexed jobId,
        address indexed provider,
        uint256 amount
    );
    
    event CollateralReturned(
        uint256 indexed jobId,
        address indexed provider,
        uint256 amount
    );
    
    event ReputationUpdated(
        address indexed user,
        uint256 newReputation,
        bool isProvider
    );

    // Modifiers
    modifier onlyRequester(uint256 _jobId) {
        require(jobs[_jobId].requester == msg.sender, "Not job requester");
        _;
    }
    
    modifier onlyProvider(uint256 _jobId) {
        require(jobs[_jobId].provider == msg.sender, "Not job provider");
        _;
    }
    
    modifier onlyVerifier() {
        require(verifiers[msg.sender], "Not authorized verifier");
        _;
    }
    
    modifier jobExists(uint256 _jobId) {
        require(_jobId > 0 && _jobId <= jobCounter, "Job does not exist");
        _;
    }

    constructor(
        address _paymentToken,
        address _priceFeed,
        address _feeRecipient
    ) {
        require(_paymentToken != address(0), "Invalid token address");
        require(_priceFeed != address(0), "Invalid price feed");
        require(_feeRecipient != address(0), "Invalid fee recipient");
        
        paymentToken = IERC20(_paymentToken);
        priceFeed = AggregatorV3Interface(_priceFeed);
        feeRecipient = _feeRecipient;
    }

    /**
     * @dev Create a new compute job
     * @param _amount Payment amount for the job
     * @param _jobHash Hash of the job specification
     * @param _deadline Deadline for job completion
     */
    function createJob(
        uint256 _amount,
        bytes32 _jobHash,
        uint256 _deadline
    ) external whenNotPaused nonReentrant returns (uint256) {
        require(_amount > 0, "Amount must be greater than 0");
        require(_jobHash != bytes32(0), "Invalid job hash");
        require(_deadline > block.timestamp, "Invalid deadline");
        
        // Transfer payment to escrow
        require(
            paymentToken.transferFrom(msg.sender, address(this), _amount),
            "Payment transfer failed"
        );
        
        jobCounter++;
        uint256 jobId = jobCounter;
        
        jobs[jobId] = Job({
            requester: msg.sender,
            provider: address(0),
            amount: _amount,
            collateral: 0,
            jobHash: _jobHash,
            resultHash: bytes32(0),
            deadline: _deadline,
            status: JobStatus.Created,
            resultVerified: false,
            createdAt: block.timestamp,
            completedAt: 0
        });
        
        requesterJobs[msg.sender].push(jobId);
        
        emit JobCreated(jobId, msg.sender, _amount, _jobHash, _deadline);
        
        return jobId;
    }

    /**
     * @dev Provider accepts and starts a job
     * @param _jobId ID of the job to accept
     */
    function acceptJob(uint256 _jobId) 
        external 
        whenNotPaused 
        nonReentrant 
        jobExists(_jobId) 
    {
        Job storage job = jobs[_jobId];
        require(job.status == JobStatus.Created, "Job not available");
        require(job.requester != msg.sender, "Cannot accept own job");
        require(block.timestamp < job.deadline, "Job deadline passed");
        
        // Calculate required collateral
        uint256 collateralAmount = job.amount.mul(collateralPercent).div(10000);
        
        // Transfer collateral from provider
        require(
            paymentToken.transferFrom(msg.sender, address(this), collateralAmount),
            "Collateral transfer failed"
        );
        
        job.provider = msg.sender;
        job.collateral = collateralAmount;
        job.status = JobStatus.Assigned;
        
        providerJobs[msg.sender].push(_jobId);
        
        emit JobAssigned(_jobId, msg.sender);
    }

    /**
     * @dev Provider submits job results
     * @param _jobId ID of the job
     * @param _resultHash Hash of the computation result
     */
    function submitResult(uint256 _jobId, bytes32 _resultHash) 
        external 
        onlyProvider(_jobId) 
        jobExists(_jobId) 
    {
        Job storage job = jobs[_jobId];
        require(job.status == JobStatus.Assigned || job.status == JobStatus.Running, "Invalid job status");
        require(_resultHash != bytes32(0), "Invalid result hash");
        require(block.timestamp <= job.deadline, "Job deadline exceeded");
        
        job.resultHash = _resultHash;
        job.status = JobStatus.Completed;
        job.completedAt = block.timestamp;
        
        emit JobCompleted(_jobId, _resultHash, block.timestamp);
    }

    /**
     * @dev Requester verifies and accepts the result
     * @param _jobId ID of the job
     */
    function verifyResult(uint256 _jobId) 
        external 
        onlyRequester(_jobId) 
        jobExists(_jobId) 
        nonReentrant 
    {
        Job storage job = jobs[_jobId];
        require(job.status == JobStatus.Completed, "Job not completed");
        require(!job.resultVerified, "Result already verified");
        
        job.resultVerified = true;
        
        // Calculate platform fee
        uint256 platformFee = job.amount.mul(platformFeePercent).div(10000);
        uint256 providerPayment = job.amount.sub(platformFee);
        
        // Transfer payment to provider
        require(
            paymentToken.transfer(job.provider, providerPayment),
            "Provider payment failed"
        );
        
        // Transfer platform fee
        require(
            paymentToken.transfer(feeRecipient, platformFee),
            "Platform fee transfer failed"
        );
        
        // Return collateral to provider
        if (job.collateral > 0) {
            require(
                paymentToken.transfer(job.provider, job.collateral),
                "Collateral return failed"
            );
            emit CollateralReturned(_jobId, job.provider, job.collateral);
        }
        
        // Update reputation
        providerReputation[job.provider] = providerReputation[job.provider].add(1);
        requesterReputation[job.requester] = requesterReputation[job.requester].add(1);
        
        emit PaymentReleased(_jobId, job.provider, providerPayment);
        emit ReputationUpdated(job.provider, providerReputation[job.provider], true);
        emit ReputationUpdated(job.requester, requesterReputation[job.requester], false);
    }

    /**
     * @dev Dispute a job result
     * @param _jobId ID of the job
     * @param _reason Reason for dispute
     */
    function disputeJob(uint256 _jobId, string memory _reason) 
        external 
        jobExists(_jobId) 
    {
        Job storage job = jobs[_jobId];
        require(
            msg.sender == job.requester || msg.sender == job.provider,
            "Not authorized to dispute"
        );
        require(
            job.status == JobStatus.Completed || job.status == JobStatus.Assigned,
            "Cannot dispute this job"
        );
        require(
            block.timestamp <= job.completedAt.add(disputeTimeout),
            "Dispute period expired"
        );
        
        job.status = JobStatus.Disputed;
        
        emit JobDisputed(_jobId, msg.sender, _reason);
    }

    /**
     * @dev Resolve a disputed job (only by authorized verifiers)
     * @param _jobId ID of the job
     * @param _requesterWins Whether the requester wins the dispute
     */
    function resolveDispute(uint256 _jobId, bool _requesterWins) 
        external 
        onlyVerifier 
        jobExists(_jobId) 
        nonReentrant 
    {
        Job storage job = jobs[_jobId];
        require(job.status == JobStatus.Disputed, "Job not in dispute");
        
        if (_requesterWins) {
            // Return payment to requester
            require(
                paymentToken.transfer(job.requester, job.amount),
                "Requester refund failed"
            );
            
            // Slash provider's collateral
            if (job.collateral > 0) {
                uint256 penalty = job.collateral.div(2);
                require(
                    paymentToken.transfer(feeRecipient, penalty),
                    "Penalty transfer failed"
                );
                
                // Return remaining collateral
                uint256 remaining = job.collateral.sub(penalty);
                if (remaining > 0) {
                    require(
                        paymentToken.transfer(job.provider, remaining),
                        "Remaining collateral transfer failed"
                    );
                }
            }
            
            // Penalize provider reputation
            if (providerReputation[job.provider] > 0) {
                providerReputation[job.provider] = providerReputation[job.provider].sub(1);
            }
        } else {
            // Provider wins - process payment normally
            verifyResult(_jobId);
        }
        
        job.status = JobStatus.Completed;
        emit JobResolved(_jobId, _requesterWins);
    }

    /**
     * @dev Cancel a job (only before it's assigned)
     * @param _jobId ID of the job
     */
    function cancelJob(uint256 _jobId) 
        external 
        onlyRequester(_jobId) 
        jobExists(_jobId) 
        nonReentrant 
    {
        Job storage job = jobs[_jobId];
        require(job.status == JobStatus.Created, "Cannot cancel assigned job");
        
        job.status = JobStatus.Cancelled;
        
        // Refund payment to requester
        require(
            paymentToken.transfer(job.requester, job.amount),
            "Refund failed"
        );
    }

    /**
     * @dev Get current price from oracle
     */
    function getLatestPrice() public view returns (int256) {
        (
            ,
            int256 price,
            ,
            ,
            
        ) = priceFeed.latestRoundData();
        return price;
    }

    /**
     * @dev Add or remove verifier (only owner)
     */
    function setVerifier(address _verifier, bool _status) external onlyOwner {
        verifiers[_verifier] = _status;
    }

    /**
     * @dev Update platform fee (only owner)
     */
    function setPlatformFee(uint256 _feePercent) external onlyOwner {
        require(_feePercent <= 1000, "Fee too high"); // Max 10%
        platformFeePercent = _feePercent;
    }

    /**
     * @dev Update collateral percentage (only owner)
     */
    function setCollateralPercent(uint256 _percent) external onlyOwner {
        require(_percent <= 5000, "Collateral too high"); // Max 50%
        collateralPercent = _percent;
    }

    /**
     * @dev Pause contract (only owner)
=======
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
>>>>>>> 4c40309e804c8f522625b7fd70da67d8d7383849
     */
    function pause() external onlyOwner {
        _pause();
    }
<<<<<<< HEAD

    /**
     * @dev Unpause contract (only owner)
=======
    
    /**
     * @notice Unpause the contract
>>>>>>> 4c40309e804c8f522625b7fd70da67d8d7383849
     */
    function unpause() external onlyOwner {
        _unpause();
    }
<<<<<<< HEAD

    /**
     * @dev Get jobs by requester
     */
    function getRequesterJobs(address _requester) external view returns (uint256[] memory) {
        return requesterJobs[_requester];
    }

    /**
     * @dev Get jobs by provider
     */
    function getProviderJobs(address _provider) external view returns (uint256[] memory) {
        return providerJobs[_provider];
=======
    
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
>>>>>>> 4c40309e804c8f522625b7fd70da67d8d7383849
    }
} 