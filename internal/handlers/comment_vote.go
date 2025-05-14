package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/phanvantai/taiphanvan_backend/internal/database"
	"github.com/phanvantai/taiphanvan_backend/internal/models"
	"gorm.io/gorm"
)

// GetUserVote retrieves a user's vote on a comment
func GetUserVote(commentID, userID uint) (models.VoteType, error) {
	var vote models.CommentVote
	result := database.DB.Where("comment_id = ? AND user_id = ?", commentID, userID).First(&vote)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return models.VoteTypeNone, nil
		}
		return models.VoteTypeNone, result.Error
	}
	return vote.VoteType, nil
}

// GetCommentVotes retrieves the vote counts for a specific comment
func GetCommentVotes(c *gin.Context) {
	// Parse comment ID from path
	commentID, err := strconv.ParseUint(c.Param("commentID"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("Invalid input", "Invalid comment ID"))
		return
	}

	// Check if comment exists
	var comment models.Comment
	if err := database.DB.First(&comment, commentID).Error; err != nil {
		c.JSON(http.StatusNotFound, models.NewErrorResponse("Not found", "Comment not found"))
		return
	}

	// Get user's current vote if authenticated
	var userVote int8 = 0
	if userID, exists := c.Get("userID"); exists {
		var vote models.CommentVote
		result := database.DB.Where("comment_id = ? AND user_id = ?", commentID, userID).First(&vote)
		if result.Error == nil {
			userVote = int8(vote.VoteType)
		}
	}

	// Return vote counts
	response := models.CommentVoteResponse{
		CommentID:   comment.ID,
		UpvoteCount: comment.UpvoteCount,
		UserVote:    userVote,
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(response, "Vote counts retrieved successfully"))
}

// VoteOnComment handles upvoting, downvoting, or removing a vote from a comment
func VoteOnComment(c *gin.Context) {
	// Get authenticated user ID
	userID, _ := c.Get("userID")

	// Parse comment ID from path
	commentID, err := strconv.ParseUint(c.Param("commentID"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("Invalid input", "Invalid comment ID"))
		return
	}

	// Check if comment exists
	var comment models.Comment
	if err := database.DB.First(&comment, commentID).Error; err != nil {
		c.JSON(http.StatusNotFound, models.NewErrorResponse("Not found", "Comment not found"))
		return
	}

	// Parse request body
	var requestBody models.CommentVoteRequest
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("Invalid input", err.Error()))
		return
	}

	// Validate vote type
	if requestBody.VoteType < -1 || requestBody.VoteType > 1 {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("Invalid input", "Invalid vote type, must be -1, 0, or 1"))
		return
	}

	// Process vote in a transaction to ensure consistency
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		// Check if user has already voted on this comment
		var existingVote models.CommentVote
		result := tx.Where("comment_id = ? AND user_id = ?", commentID, userID).First(&existingVote)

		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				// User hasn't voted before, create a new vote
				if requestBody.VoteType != models.VoteTypeNone {
					newVote := models.CommentVote{
						UserID:    userID.(uint),
						CommentID: uint(commentID),
						VoteType:  requestBody.VoteType,
					}
					if err := tx.Create(&newVote).Error; err != nil {
						return err
					}

					// Update comment vote count
					voteChange := int(requestBody.VoteType)
					if err := tx.Model(&comment).UpdateColumn("upvote_count", gorm.Expr("upvote_count + ?", voteChange)).Error; err != nil {
						return err
					}
				}
			} else {
				// Database error
				return result.Error
			}
		} else {
			// User has already voted, handle vote change
			oldVoteType := existingVote.VoteType

			if requestBody.VoteType == models.VoteTypeNone {
				// Remove vote
				if err := tx.Delete(&existingVote).Error; err != nil {
					return err
				}

				// Update comment vote count
				voteChange := -int(oldVoteType)
				if err := tx.Model(&comment).UpdateColumn("upvote_count", gorm.Expr("upvote_count + ?", voteChange)).Error; err != nil {
					return err
				}
			} else if oldVoteType != requestBody.VoteType {
				// Change vote
				existingVote.VoteType = requestBody.VoteType
				if err := tx.Save(&existingVote).Error; err != nil {
					return err
				}

				// Calculate vote change and update comment vote count
				voteChange := int(requestBody.VoteType - oldVoteType)
				if err := tx.Model(&comment).UpdateColumn("upvote_count", gorm.Expr("upvote_count + ?", voteChange)).Error; err != nil {
					return err
				}
			}
			// If the vote didn't change, do nothing
		}

		// Reload comment to get updated vote count
		return tx.First(&comment, commentID).Error
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse("Server error", "Failed to process vote: "+err.Error()))
		return
	}

	// Return updated vote counts
	response := models.CommentVoteResponse{
		CommentID:   comment.ID,
		UpvoteCount: comment.UpvoteCount,
		UserVote:    int8(requestBody.VoteType),
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(response, "Vote processed successfully"))
}
