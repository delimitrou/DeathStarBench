package com.idugalic.domain;

import java.util.Date;

import org.springframework.data.annotation.CreatedBy;
import org.springframework.data.annotation.CreatedDate;
import org.springframework.data.annotation.LastModifiedBy;
import org.springframework.data.annotation.LastModifiedDate;
import org.springframework.data.mongodb.core.mapping.Document;

@Document
@SuppressWarnings("serial")
public abstract class BaseAuditableEntity extends BaseEntity {

	@CreatedBy
	private String createdBy;

	@CreatedDate
	private Date createdTime;

	@LastModifiedBy
	private String lastModifiedBy;

	@LastModifiedDate
	private Date lastModifiedTime;

	public String getCreatedBy() {
		return createdBy;
	}

	public Date getCreatedTime() {
		return createdTime;
	}

	public String getLastModifiedBy() {
		return lastModifiedBy;
	}

	public Date getLastModifiedTime() {
		return lastModifiedTime;
	}

}
