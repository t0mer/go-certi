package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/t0mer/go-certi/internal/models"
)

func certToResponse(cert models.Certificate) CertificateResponse {
	var sans []string
	json.Unmarshal([]byte(cert.Sans), &sans) //nolint:errcheck
	if sans == nil {
		sans = []string{}
	}
	return CertificateResponse{
		ID:           cert.ID,
		FQDNID:       cert.FqdnID,
		Serial:       cert.Serial,
		IssuerCA:     cert.IssuerCa,
		SubjectCN:    cert.SubjectCn,
		SANs:         sans,
		NotBefore:    cert.NotBefore,
		NotAfter:     cert.NotAfter,
		DiscoveredAt: cert.DiscoveredAt,
		Source:       cert.Source,
	}
}

// ListCertificates godoc
// @Summary List certificates with optional filtering and pagination
// @Tags certificates
// @Produce json
// @Param fqdn query string false "Filter by FQDN hostname"
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Page size (default: 25, max: 200)"
// @Success 200 {object} CertificateListResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/certificates [get]
func (h *Handler) ListCertificates(c *gin.Context) {
	ctx := c.Request.Context()
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "25"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 25
	}
	offset := int64((page - 1) * pageSize)
	limit := int64(pageSize)

	fqdnName := c.Query("fqdn")
	var items []models.Certificate
	var total int64
	var err error

	if fqdnName != "" {
		f, ferr := h.q.GetFQDNByName(ctx, fqdnName)
		if ferr != nil {
			c.JSON(http.StatusOK, CertificateListResponse{Items: []CertificateResponse{}, Total: 0, Page: page, PageSize: pageSize})
			return
		}
		items, err = h.q.ListCertificatesByFQDN(ctx, models.ListCertificatesByFQDNParams{
			FqdnID: f.ID, Limit: limit, Offset: offset,
		})
		total, _ = h.q.CountCertificatesByFQDN(ctx, f.ID)
	} else {
		items, err = h.q.ListCertificates(ctx, models.ListCertificatesParams{
			Limit: limit, Offset: offset,
		})
		total, _ = h.q.CountCertificates(ctx)
	}
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	resp := make([]CertificateResponse, len(items))
	for i, cert := range items {
		resp[i] = certToResponse(cert)
	}
	c.JSON(http.StatusOK, CertificateListResponse{
		Items: resp, Total: total, Page: page, PageSize: pageSize,
	})
}

// GetCertificate godoc
// @Summary Get a single certificate by ID
// @Tags certificates
// @Produce json
// @Param id path string true "Certificate ID"
// @Success 200 {object} CertificateResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/certificates/{id} [get]
func (h *Handler) GetCertificate(c *gin.Context) {
	cert, err := h.q.GetCertificate(c.Request.Context(), c.Param("id"))
	if err != nil {
		notFoundOr500(c, err)
		return
	}
	c.JSON(http.StatusOK, certToResponse(cert))
}

// ListCAs godoc
// @Summary List distinct certificate authorities seen in CT logs
// @Tags certificates
// @Produce json
// @Success 200 {array} string
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/certificates/cas [get]
func (h *Handler) ListCAs(c *gin.Context) {
	cas, err := h.q.ListDistinctCAs(c.Request.Context())
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	if cas == nil {
		cas = []string{}
	}
	c.JSON(http.StatusOK, cas)
}
