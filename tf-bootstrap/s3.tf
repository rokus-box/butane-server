resource "aws_s3_bucket" "terraform-lock" {
  bucket = "asdfnq9wfjehjfdajcn98e"
}

resource "aws_s3_bucket_ownership_controls" "terraform-lock" {
  bucket = aws_s3_bucket.terraform-lock.id
  rule {
    object_ownership = "BucketOwnerPreferred"
  }
}

resource "aws_s3_bucket_acl" "terraform-lock" {
  depends_on = [aws_s3_bucket_ownership_controls.terraform-lock]

  bucket = aws_s3_bucket.terraform-lock.id
  acl    = "private"
}