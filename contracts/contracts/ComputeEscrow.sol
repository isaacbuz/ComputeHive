// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/security/ReentrancyGuard.sol";
import "@openzeppelin/contracts/security/Pausable.sol";
import "@openzeppelin/contracts/utils/math/SafeMath.sol";
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@chainlink/contracts/src/v0.8/interfaces/AggregatorV3Interface.sol";

/**
 * @title ComputeEscrow
 * @dev Main escrow contract for ComputeHive distributed compute jobs
 */
contract ComputeEscrow is Ownable, ReentrancyGuard, Pausable {
    using SafeMath for uint256;

    // Job states
    enum JobStatus {
        Created,
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
    }

    // State variables
    mapping(uint256 => Job) public jobs;
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
     */
    function pause() external onlyOwner {
        _pause();
    }

    /**
     * @dev Unpause contract (only owner)
     */
    function unpause() external onlyOwner {
        _unpause();
    }

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
    }
} 