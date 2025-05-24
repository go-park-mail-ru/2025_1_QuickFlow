drop trigger if exists trg_update_post_comment_count ON comment;
drop FUNCTION if exists update_post_comment_count();

drop trigger if exists trg_update_comment_like_count ON like_comment;
