package handlers

import (
	"fmt"
)

func generateFAQText() string {
	return "<b>📋 Frequently Asked Questions</b>\n\n" +

		"<b>🔍 Course Information</b>\n" +
		"❓ <b>How do I check a course?</b>\n" +
		"   Simply send a course code (e.g., <b>PHYS 161</b>, <b>CSCI 151</b>) without any command. The bot will show current enrollment and section details.\n\n" +

		"<b>🚨 Troubleshooting</b>\n" +
		"❓ <b>What if a course is not found?</b>\n" +
		"   • Check the course code spelling\n" +
		"   • Ensure the course is offered this semester\n" +

		"❓ <b>What if a section is not found?</b>\n" +
		"   • Check section naming (1L, 2PLB, 3R, etc.)\n" +
		"   • Make sure section number comes first (1L, 2PLB, 3R, etc.)\n" +
		"   • Ensure the section exists for that course\n" +

		"❓ <b>Bot not responding?</b>\n" +
		"   • Wait a moment and try again\n" +
		"   • Use <code>/start</code> to reset your session\n" +
		"   • Check your internet connection\n\n" +

		"❓ <b>Not getting notifications?</b>\n" +
		"   • Verify your subscription with <code>/list</code>\n" +
		"   • Ensure you haven't blocked the bot\n" +
		"   • Notifications only come when spots open up\n\n" +
		"❓ <b>How often does the bot update course data?</b>\n" +
		"   • The bot uses a dynamic schedule to check for updates more frequently as registration deadlines approach.\n" +
		"   • Default time is 3 hour frequency\n" +
		"   • From 1 hour to 30 minutes before registration closes: updates every 30 minutes.\n" +
		"   • From 30 to 15 minutes before: updates every 15 minutes.\n" +
		"   • From 15 to 5 minutes before: updates every 5 minutes.\n" +
		"   • In the last 5 minutes before registration closes and 5 minutes after: updates every minute.\n" +
		"   • From 5 to 30 minutes after registration closes: updates every 3 minutes.\n" +
		"   • This ensures you get the most up-to-date information when it matters most!\n\n" +

		"<b>💰 Support</b>\n" +
		"❓ <b>How can I support this bot?</b>\n" +
		"   Use <code>/donate</code> to see donation information. Your support helps maintain the bot and improve its features.\n\n"
}

func generateWelcomeText(semester string) string {
	return fmt.Sprintf(
		"<b>Welcome to the NU Course Info.</b> 🎓\n\n"+
			"I provide real-time insights about class enrollments for <b>%s</b>\n\n"+
			"Simply send me a course code (e.g. <b>PHYS 161</b>) to get:\n"+
			"• Current enrollment numbers\n"+
			"• Available seats\n"+
			"• Section details\n\n"+
			"Also provides opportunity to track course status by subscription system with notifications\n\n"+
			"<i>Updates every 60/30/15/5 minutes \n[The closer to registration the more frequent updates will be]</i>",
		semester)
}
