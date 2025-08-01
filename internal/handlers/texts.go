package handlers

import (
	"fmt"
)

func generateFAQText() string {
	return "*📋 Frequently Asked Questions*\n\n" +

		"*🔍 Course Information*\n" +
		"❓ *How do I check a course?*\n" +
		"   Simply send a course code \\(e\\.g\\., *PHYS 161*, *CSCI 151*\\) without any command\\. The bot will show current enrollment and section details\\.\n\n" +

		"*🚨 Troubleshooting*\n" +
		"❓ *What if a course is not found?*\n" +
		"   • Check the course code spelling\n" +
		"   • Ensure the course is offered this semester\n" +

		"❓ *What if a section is not found?*\n" +
		"   • Check section naming \\(1L, 2PLB, 3R, etc\\.\\)\n" +
		"   • Make sure section number comes first \\(1L, 2PLB, 3R, etc\\.\\)\n" +
		"   • Ensure the section exists for that course\n" +

		"❓ *Bot not responding?*\n" +
		"   • Wait a moment and try again\n" +
		"   • Use `/start` to reset your session\n" +
		"   • Check your internet connection\n\n" +

		"❓ *Not getting notifications?*\n" +
		"   • Verify your subscription with `/list`\n" +
		"   • Ensure you haven't blocked the bot\n" +
		"   • Notifications only come when spots open up\n\n" +

		"*💰 Support*\n" +
		"❓ *How can I support this bot?*\n" +
		"   Use `/donate` to see donation information\\. Your support helps maintain the bot and improve its features\\.\n\n"
}

func generateWelcomeText(semester string) string {
	return fmt.Sprintf(
		"*Welcome to the NU Course Info\\.* 🎓\n\n"+
			"I provide real\\-time insights about class enrollments for *%s*\n\n"+
			"Simply send me a course code \\(e\\.g\\. *PHYS 161*\\) to get:\n"+
			"• Current enrollment numbers\n"+
			"• Available seats\n"+
			"• Section details\n\n"+
			"Also provides opportunity to track course status by subscription system with notifications\n\n"+
			"_Updates every 60/30/15/5 minutes \n\\[The closer to registration the more frequent updates will be\\]_",
		semester)
}
