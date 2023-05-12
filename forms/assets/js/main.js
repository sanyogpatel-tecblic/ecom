$(window).scroll(function() {    
    var scroll = $(window).scrollTop();

    if (scroll >= 500) {
        $(".navbar").addClass("active");
    } else {
        $(".navbar").removeClass("active");
    }
});