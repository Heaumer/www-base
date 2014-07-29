$(document).ready(function(){
	/* hacky, hacky */
	function mkcontent(c) {
		c = String(c).replace(/<br[^>]*>/g, "\n")
		c = c.replace(/<[^>]+>/g, "")
		return c
	}

	$('form').each(function () {
		$(this).submit(function() {
			$(this).find("input[name=Content]").val(
				mkcontent($(this).find("div[name=content]").html()))
			return true
		})
	})
})