$(".qor-seo-submit").click(function(){
  var fieldName = $(".qor-seo-field").attr("name");
  var $wrap = $(this).parents(".qor-seo");
  var titleValue = $wrap.find(".qor-seo-title-field").val();
  var descriptionValue = $wrap.find(".qor-seo-description-field").val();
  var data = {};
  data[fieldName] = "{ \"Title\" : \"" + titleValue + "\", \"Description\" : \"" + descriptionValue + "\"}";
  console.info(data);
  $.ajax({
    type: "POST",
    url: "/admin/seo/1",
    data: data,
    success: function () {
      alert("Save");
    },
    error: function (data) {
      alert("Can't save");
    }
  });
  return false;
});
